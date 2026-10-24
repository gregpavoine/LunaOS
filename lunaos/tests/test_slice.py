import concurrent.futures
import base64
import hashlib
import hmac
import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import tempfile
import threading
import time
import unittest

from luna.engine import Engine, Invalid, Unauthorized
from luna.policy import Broker, UnixConnection

ROOT = Path(__file__).resolve().parents[1]


class SliceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.build = tempfile.TemporaryDirectory(prefix='luna-build-')
        cls.binary = str(Path(cls.build.name) / 'broker')
        subprocess.run(['go', 'build', '-buildvcs=false', '-o', cls.binary, './services/policy'], cwd=ROOT, check=True)

    @classmethod
    def tearDownClass(cls):
        cls.build.cleanup()

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix='luna-')
        self.root = Path(self.tmp.name)
        self.root.chmod(0o700)
        self.key = self.root / 'owner.key'
        self.key.write_text('a' * 64)
        self.key.chmod(0o600)
        self.sock = str(self.root / 'broker.sock')
        self.proc = subprocess.Popen([self.binary, '--socket', self.sock, '--key-file', str(self.key)], stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
        for _ in range(100):
            if Path(self.sock).exists():
                break
            if self.proc.poll() is not None:
                self.fail(self.proc.stderr.read().decode())
            time.sleep(.02)
        self.db = self.root / 'goals.sqlite'
        self.broker = Broker(self.sock, self.key.read_text())
        self.engine = Engine(self.db, self.broker)

    def tearDown(self):
        self.proc.terminate()
        self.proc.wait(timeout=5)
        self.proc.stderr.close()
        self.tmp.cleanup()

    def create(self, **kw):
        return self.engine.create({'action': 'local.checksum', 'metrics': dict.fromkeys(['impact', 'urgency', 'user_value', 'confidence', 'efficiency'], 1), **kw}, 'a' * 64, kw.pop('key', 'request-1'))

    def test_owner_only_privileges(self):
        for credential in ['', 'guest', 'ok je valide']:
            with self.assertRaises(Unauthorized):
                self.engine.create({'action': 'local.checksum'}, credential, 'unauth')
        self.assertEqual(self.engine.list_goals(), [])
        self.assertFalse(Broker(self.sock, 'guest').issue('g', 'local.checksum')['allowed'])

    def test_absent_budget_payment_refused(self):
        goal = self.create(action='payment.send')
        self.assertEqual(goal['status'], 'blocked')
        self.assertEqual(goal['decision_code'], 'BUDGET_UNSET')
        self.assertIsNone(goal['computed_priority'])
        self.assertEqual(self.engine.tick()['outcome'], 'idle_success')

    def test_policy_deny_before_score(self):
        goal = self.create(action='system.root')
        self.assertEqual(goal['decision_code'], 'ACTION_DENIED')
        self.assertIsNone(goal['computed_priority'])
        self.assertNotIn('goal.scored', [e['event_type'] for e in self.engine.events(goal['id'])])

    def test_duplicate_job_idempotency(self):
        first = self.create()
        self.assertEqual(first['id'], self.create()['id'])
        with self.assertRaises(Invalid):
            self.create(action='system.root')
        with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
            results = list(pool.map(lambda _: Engine(self.db, self.broker).tick(), range(2)))
        self.assertEqual(sum(r['outcome'] == 'succeeded' for r in results), 1)
        goal = self.engine.get(first['id'])
        self.assertEqual(goal['status'], 'completed')
        self.assertEqual(len(goal['evidence']), 1)
        self.assertEqual(goal['evidence'][0]['sha256'], hashlib.sha256(b'LunaOS public checkpoint v1\n').hexdigest())
        self.assertEqual(self.engine.delivery_counts(), (1, 1))

    def test_unknown_source_is_hypothesis(self):
        canary = 'ASSERTION_TEXT_MUST_NOT_BE_STORED'
        goal = self.create(assertion={'claim': canary, 'source': None, 'confidence': 1})
        self.assertEqual(goal['assertion']['provenance_status'], 'unknown')
        self.assertEqual(goal['assertion']['epistemic_status'], 'hypothesis')
        self.assertNotIn(canary.encode(), self.db.read_bytes())
        self.engine.tick()
        self.assertEqual(self.engine.get(goal['id'])['assertion']['epistemic_status'], 'hypothesis')

    def test_inactive_idle_success(self):
        self.assertEqual(self.engine.tick(), {'outcome': 'idle_success', 'reason': 'NO_AUTHORIZED_USEFUL_WORK'})
        self.create(metrics=dict.fromkeys(['impact', 'urgency', 'user_value', 'confidence', 'efficiency'], 0))
        self.assertEqual(self.engine.tick()['outcome'], 'idle_success')
        self.assertEqual(len(self.engine.list_goals()), 1)

    def test_queued_goal_resumes_in_new_process(self):
        goal = self.create()
        result = subprocess.run([sys.executable, '-m', 'luna', 'tick', '--db', str(self.db), '--broker-socket', self.sock, '--key-file', str(self.key)], cwd=ROOT, text=True, capture_output=True, check=True)
        self.assertEqual(json.loads(result.stdout)['outcome'], 'succeeded')
        self.assertEqual(self.engine.get(goal['id'])['status'], 'completed')

    def test_broker_outage_cannot_execute_queued_work(self):
        goal = self.create()
        self.proc.terminate()
        self.proc.wait(timeout=5)
        self.assertEqual(self.engine.tick()['outcome'], 'idle_success')
        self.assertEqual(self.engine.get(goal['id'])['status'], 'approved')
        self.assertEqual(self.engine.get(goal['id'])['evidence'], [])

    def test_forged_grant_and_sensitive_domain_denied(self):
        goal = self.create(privacy_class='intimate')
        self.assertEqual(goal['status'], 'blocked')
        self.assertFalse(self.broker.evaluate('g', 'local.checksum', 'forged')['allowed'])

    def test_score_reproducibility_and_ranking(self):
        low = self.create(metrics=dict.fromkeys(['impact', 'urgency', 'user_value', 'confidence', 'efficiency'], .2))
        high = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'high')
        self.assertEqual(low['computed_priority'], 20)
        self.assertEqual(high['computed_priority'], 100)
        self.assertEqual(self.engine.tick()['goal_id'], high['id'])

    def test_input_validation(self):
        for body in [{'action': 'local.checksum', 'metrics': {'impact': float('nan')}}, {'action': 'local.checksum', 'owner_id': 'guest'}, {'action': 'local.checksum', 'assertion': {'claim': 'x', 'provenance_status': 'verified'}}]:
            with self.assertRaises(Invalid):
                self.engine.create(body, 'a' * 64, 'bad')


    def test_kill_during_execution_rolls_back_then_resumes(self):
        goal = self.create()
        program = """
import os, signal, sys
from luna.engine import Engine
from luna.policy import Broker
import luna.engine as module
original = module.execute
def killed(action):
    result = original(action)
    os.kill(os.getpid(), signal.SIGKILL)
    return result
module.execute = killed
Engine(sys.argv[1], Broker(sys.argv[2], 'a'*64)).tick()
"""
        killed = subprocess.run([sys.executable, '-c', program, str(self.db), self.sock], cwd=ROOT, capture_output=True)
        self.assertEqual(killed.returncode, -9)
        after = self.engine.get(goal['id'])
        self.assertEqual(after['execution']['execution_state'], 'queued')
        self.assertEqual(after['evidence'], [])
        self.assertEqual(self.engine.tick()['outcome'], 'succeeded')
        self.assertEqual(len(self.engine.get(goal['id'])['evidence']), 1)

    def test_api_owner_auth_and_real_round_trip(self):
        api_socket = str(self.root / 'api.sock')
        server = subprocess.Popen([sys.executable, '-m', 'luna', 'serve', '--db', str(self.db), '--broker-socket', self.sock, '--key-file', str(self.key), '--api-socket', api_socket], cwd=ROOT, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        try:
            for _ in range(100):
                if Path(api_socket).exists():
                    break
                time.sleep(.02)
            def request(method, path, body=None, token='a'*64):
                conn = UnixConnection(api_socket)
                try:
                    conn.request(method, path, json.dumps(body) if body is not None else None, {'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json', 'Idempotency-Key': 'api-request'})
                    response = conn.getresponse()
                    return response.status, json.loads(response.read())
                finally:
                    conn.close()
            for token in ['', 'guest', 'ok je valide']:
                self.assertEqual(request('POST', '/v1/goals', {'action': 'local.checksum'}, token)[0], 401)
                self.assertEqual(request('GET', '/v1/goals', token=token)[0], 401)
            status, goal = request('POST', '/v1/goals', {'action': 'local.checksum'})
            self.assertEqual(status, 201)
            self.assertNotIn('grant_token', goal)
            self.assertEqual(request('POST', '/v1/scheduler/tick', {})[1]['outcome'], 'succeeded')
            self.assertEqual(request('GET', '/v1/goals/' + goal['id'])[1]['status'], 'completed')
            self.assertEqual(request('GET', '/v1/goals/' + goal['id'] + '/decisions')[0], 200)
            planned = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'api-plan-goal')
            plan_path = '/v1/goals/' + planned['id'] + '/plan'
            self.assertEqual(request('POST', plan_path, {'depends_on': [goal['id']], 'parent_id': goal['id']}, token='guest')[0], 401)
            plan_status, plan_goal = request('POST', plan_path, {'depends_on': [goal['id']], 'parent_id': goal['id']})
            self.assertEqual(plan_status, 200)
            self.assertEqual(plan_goal['plan'], {'parent_id': goal['id'], 'depends_on': [goal['id']]})
            self.assertEqual(request('POST', plan_path, {'depends_on': [], 'unexpected': True})[0], 400)
            self.assertEqual(request('POST', '/v1/goals/missing/plan', {'depends_on': []})[0], 404)
            self.assertEqual(request('GET', '/v1/system/capabilities')[1]['remote'], 'disabled')
            self.assertEqual(request('GET', '/v1/not-implemented')[0], 404)
            self.assertEqual(os.stat(api_socket).st_mode & 0o777, 0o600)
        finally:
            server.terminate()
            server.wait(timeout=5)
            server.stdout.close()
            server.stderr.close()

    def test_private_database_and_append_only_events(self):
        goal = self.create()
        self.assertEqual(self.db.stat().st_mode & 0o777, 0o600)
        import sqlite3
        with self.engine.store.connect() as db:
            with self.assertRaises(sqlite3.IntegrityError):
                db.execute('DELETE FROM goal_event')
            self.assertEqual(db.execute('PRAGMA integrity_check').fetchone()[0], 'ok')
            self.assertEqual(db.execute('SELECT count(*) FROM schema_migration').fetchone()[0], 2)
        Engine(self.db, self.broker)
        self.assertEqual(self.engine.get(goal['id'])['status'], 'approved')

    def test_store_migrates_v1_without_data_loss(self):
        import sqlite3
        from luna.store import Store
        legacy = self.root / 'legacy.sqlite'
        conn = sqlite3.connect(legacy)
        conn.executescript((ROOT / 'migrations/core/001_bootstrap.sql').read_text())
        conn.execute("INSERT INTO goal(id,owner_id,action,privacy_class,status,metrics,computed_priority,priority_explanation,decision_code,assertion,grant_token,created_at,updated_at) VALUES('legacy-goal','owner','local.checksum','general','completed','{}',100,'{}','OK',NULL,'redacted','t0','t1')")
        conn.execute("INSERT INTO goal_execution(id,goal_id,idempotency_key,request_hash,execution_state) VALUES('legacy-run','legacy-goal','legacy-key','hash','succeeded')")
        conn.execute("INSERT INTO evidence(id,goal_id,sha256,byte_count,provenance,verified_at) VALUES('legacy-evidence','legacy-goal','aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',1,'fixture','t1')")
        conn.execute("INSERT INTO goal_event(event_id,goal_id,event_type,decision_code,correlation_id,occurred_at) VALUES('legacy-event','legacy-goal','goal.completed','OK','legacy-goal','t1')")
        conn.execute("INSERT INTO event_outbox(event_id,delivered_at) VALUES('legacy-event','t1')")
        conn.commit()
        conn.close()
        legacy.chmod(0o600)
        Store(legacy)
        with Store(legacy).connect() as upgraded:
            self.assertEqual(upgraded.execute('SELECT count(*) FROM schema_migration').fetchone()[0], 2)
            self.assertEqual(upgraded.execute("SELECT id FROM goal WHERE id='legacy-goal'").fetchone()[0], 'legacy-goal')
            self.assertEqual(upgraded.execute("SELECT id FROM evidence WHERE id='legacy-evidence'").fetchone()[0], 'legacy-evidence')
            self.assertEqual(upgraded.execute("SELECT event_id FROM goal_event WHERE event_id='legacy-event'").fetchone()[0], 'legacy-event')
            self.assertEqual(upgraded.execute("SELECT parent_id FROM goal WHERE id='legacy-goal'").fetchone()[0], None)

    def test_concurrent_store_initialization_applies_migrations_once(self):
        from luna.store import Store
        database = self.root / 'concurrent.sqlite'
        barrier = threading.Barrier(2)
        def initialize(_):
            barrier.wait(timeout=5)
            return Store(database)
        with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
            list(pool.map(initialize, range(2)))
        with Store(database).connect() as db:
            self.assertEqual(db.execute('SELECT count(*) FROM schema_migration').fetchone()[0], 2)

    def test_goal_plan_persists_replaces_dependencies_and_parent(self):
        root = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'plan-root')
        dep_a = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'plan-dep-a')
        dep_b = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'plan-dep-b')
        plan = self.engine.set_plan(root['id'], [dep_a['id']], dep_b['id'], 'a' * 64)
        self.assertEqual(plan, {'parent_id': dep_b['id'], 'depends_on': [dep_a['id']]})
        self.assertEqual(self.engine.get(root['id'])['plan'], plan)
        replacement = self.engine.set_plan(root['id'], [dep_b['id']], None, 'a' * 64)
        self.assertEqual(replacement, {'parent_id': None, 'depends_on': [dep_b['id']]})

    def test_goal_plan_rejects_invalid_and_cyclic_graph_atomically(self):
        a = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'cycle-a')
        b = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'cycle-b')
        with self.assertRaises(Invalid):
            self.engine.set_plan(a['id'], [a['id']], None, 'a' * 64)
        with self.assertRaises(Invalid):
            self.engine.set_plan(a['id'], ['missing-goal'], None, 'a' * 64)
        with self.assertRaises(Invalid):
            self.engine.set_plan(a['id'], [b['id'], b['id']], None, 'a' * 64)
        with self.assertRaisesRegex(Invalid, 'PLAN_TOO_LARGE'):
            self.engine.set_plan(a['id'], [f'goal-{index}' for index in range(33)], None, 'a' * 64)
        self.engine.set_plan(a['id'], [b['id']], None, 'a' * 64)
        with self.assertRaises(Invalid):
            self.engine.set_plan(b['id'], [a['id']], None, 'a' * 64)
        self.assertEqual(self.engine.get(b['id'])['plan'], {'parent_id': None, 'depends_on': []})

    def test_goal_plan_requires_owner_and_rejects_parent_cycle(self):
        a = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'parent-a')
        b = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'parent-b')
        with self.assertRaises(Unauthorized):
            self.engine.set_plan(a['id'], [], b['id'], 'guest')
        self.engine.set_plan(a['id'], [], b['id'], 'a' * 64)
        with self.assertRaises(Invalid):
            self.engine.set_plan(b['id'], [], a['id'], 'a' * 64)

    def test_goal_plan_rejects_already_blocked_prerequisite(self):
        root = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'blocked-prereq-root')
        blocked = self.engine.create({'action': 'system.root'}, 'a' * 64, 'blocked-prereq')
        self.assertEqual(blocked['status'], 'blocked')
        with self.assertRaisesRegex(Invalid, 'DEPENDENCY_NOT_RUNNABLE'):
            self.engine.set_plan(root['id'], [blocked['id']], None, 'a' * 64)

    def test_expired_grant_blocks_goal_for_owner_replacement(self):
        goal = self.create()
        payload = json.dumps({'goal_id': goal['id'], 'action': 'local.checksum', 'principal': 'owner', 'expires': 1}, separators=(',', ':')).encode()
        encoded = base64.urlsafe_b64encode(payload).decode().rstrip('=')
        signature = hmac.new(('a' * 64).encode(), payload, hashlib.sha256).hexdigest()
        expired_grant = encoded + '.' + signature
        with self.engine.store.transaction() as db:
            db.execute('UPDATE goal SET grant_token=? WHERE id=?', (expired_grant, goal['id']))
        self.assertEqual(self.engine.tick()['outcome'], 'idle_success')
        blocked = self.engine.get(goal['id'])
        self.assertEqual(blocked['status'], 'blocked')
        self.assertEqual(blocked['decision_code'], 'GRANT_RENEWAL_REQUIRED')
        self.assertEqual(blocked['execution']['execution_state'], 'blocked')
        self.assertIn(('goal.authorization.expired', 'GRANT_RENEWAL_REQUIRED'), [(e['event_type'], e['decision_code']) for e in self.engine.events(goal['id'])])

    def test_sql_trigger_enforces_dependency_limit_on_update(self):
        import uuid
        ids = [str(uuid.uuid4()) for _ in range(35)]
        target, source, *others = ids
        prerequisites = others[:32]
        extra = others[32]
        with self.engine.store.transaction() as db:
            for goal_id in ids:
                db.execute("INSERT INTO goal(id,owner_id,action,privacy_class,status,metrics,computed_priority,priority_explanation,decision_code,assertion,grant_token,created_at,updated_at) VALUES(?,'owner','local.checksum','general','approved','{}',50,'{}','OK',NULL,'redacted','t0','t0')", (goal_id,))
            db.executemany("INSERT INTO goal_dependency(goal_id,depends_on_goal_id,dependency_type,created_at) VALUES(?,?,'blocks','t0')", [(target, item) for item in prerequisites])
            db.execute("INSERT INTO goal_dependency(goal_id,depends_on_goal_id,dependency_type,created_at) VALUES(?,?,'blocks','t0')", (source, extra))
        import sqlite3
        with self.assertRaises(sqlite3.IntegrityError):
            with self.engine.store.transaction() as db:
                db.execute('UPDATE goal_dependency SET goal_id=? WHERE goal_id=?', (target, source))

    def test_scheduler_waits_for_blocking_goal_to_complete(self):
        dependent = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'dependent')
        prerequisite = self.engine.create({'action': 'local.checksum'}, 'a' * 64, 'prerequisite')
        self.engine.set_plan(dependent['id'], [prerequisite['id']], None, 'a' * 64)
        first = self.engine.tick()
        self.assertEqual(first['goal_id'], prerequisite['id'])
        self.assertEqual(self.engine.get(dependent['id'])['status'], 'approved')
        second = self.engine.tick()
        self.assertEqual(second['goal_id'], dependent['id'])


    def test_denied_sensitive_payload_not_persisted(self):
        canary = 'PRIVATE_TEST_CANARY_DO_NOT_PERSIST'
        goal = self.create(privacy_class='intimate', assertion={'claim': canary})
        self.assertEqual(goal['status'], 'blocked')
        self.assertIsNone(goal['assertion'])
        self.assertNotIn(canary.encode(), self.db.read_bytes())

    def test_repeated_broker_outage_does_not_flood_events(self):
        goal = self.create()
        self.proc.terminate()
        self.proc.wait(timeout=5)
        self.engine.tick()
        count = len(self.engine.events(goal['id']))
        self.engine.tick()
        self.assertEqual(len(self.engine.events(goal['id'])), count)


if __name__ == '__main__':
    unittest.main()
