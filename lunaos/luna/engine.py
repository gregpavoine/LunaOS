"""Policy-first local goal engine. Transactions serialize pure actions only."""
from datetime import datetime, timezone
import hashlib
import hmac
import json
import math
import re
import sqlite3
import uuid

from .actions import execute, verify
from .store import Store

WEIGHTS = {'impact': .30, 'urgency': .20, 'user_value': .20, 'confidence': .15, 'efficiency': .15}


class Invalid(ValueError):
    pass


class Unauthorized(PermissionError):
    pass


def now():
    return datetime.now(timezone.utc).isoformat()


def uid():
    return str(uuid.uuid4())


def normalize(body):
    if not isinstance(body, dict) or set(body) - {'action', 'metrics', 'privacy_class', 'assertion'}:
        raise Invalid('INVALID_FIELDS')
    action = body.get('action')
    if not isinstance(action, str) or not re.fullmatch(r'[a-z][a-z0-9_.]{0,79}', action):
        raise Invalid('INVALID_ACTION')
    privacy = body.get('privacy_class', 'general')
    if privacy not in ('general', 'intimate', 'biometric'):
        raise Invalid('INVALID_PRIVACY_CLASS')
    supplied = body.get('metrics', {})
    if not isinstance(supplied, dict) or set(supplied) - WEIGHTS.keys():
        raise Invalid('INVALID_METRICS')
    metrics = {key: supplied.get(key, 1) for key in WEIGHTS}
    if any(type(v) not in (int, float) or not math.isfinite(v) or not 0 <= v <= 1 for v in metrics.values()):
        raise Invalid('METRICS_MUST_BE_FINITE_0_TO_1')
    assertion = body.get('assertion')
    if assertion is not None:
        if not isinstance(assertion, dict) or set(assertion) - {'claim', 'source', 'confidence'}:
            raise Invalid('INVALID_ASSERTION')
        claim, source = assertion.get('claim'), assertion.get('source')
        confidence = assertion.get('confidence', 0)
        if not isinstance(claim, str) or not 1 <= len(claim) <= 512 or source is not None and (not isinstance(source, str) or len(source) > 256):
            raise Invalid('INVALID_ASSERTION')
        if type(confidence) not in (int, float) or not math.isfinite(confidence) or not 0 <= confidence <= 1:
            raise Invalid('INVALID_CONFIDENCE')
        # The text is validated but never retained or hashed: this slice stores
        # only epistemic/provenance metadata, not user-authored assertion content.
        assertion = {'confidence': confidence, 'provenance_status': 'unknown', 'epistemic_status': 'hypothesis'}
    if privacy != 'general':
        assertion = None  # No sensitive payload or payload hash enters this store.
    return {'action': action, 'privacy_class': privacy, 'metrics': metrics, 'assertion': assertion}


def score(metrics):
    return round(100 * sum(WEIGHTS[k] * metrics[k] for k in WEIGHTS), 6)


class Engine:
    def __init__(self, path, broker):
        self.store, self.broker = Store(path), broker

    def authenticate(self, token):
        if not isinstance(token, str) or not hmac.compare_digest(token, self.broker.token):
            raise Unauthorized('OWNER_AUTH_REQUIRED')

    @staticmethod
    def event(db, goal_id, kind, code):
        event_id = uid()
        db.execute('INSERT INTO goal_event(event_id,goal_id,event_type,decision_code,correlation_id,occurred_at) VALUES(?,?,?,?,?,?)', (event_id, goal_id, kind, code, goal_id, now()))
        db.execute('INSERT INTO event_outbox(event_id) VALUES(?)', (event_id,))
        return event_id

    def create(self, body, token, idempotency_key):
        self.authenticate(token)
        if not isinstance(idempotency_key, str) or not re.fullmatch(r'[A-Za-z0-9_.:-]{1,128}', idempotency_key):
            raise Invalid('IDEMPOTENCY_KEY_REQUIRED')
        data = normalize(body)
        request_hash = hashlib.sha256(json.dumps(data, sort_keys=True).encode()).hexdigest()
        with self.store.transaction() as db:
            old = db.execute('SELECT goal_id,request_hash FROM goal_execution WHERE idempotency_key=?', (idempotency_key,)).fetchone()
            if old:
                if old['request_hash'] != request_hash:
                    raise Invalid('IDEMPOTENCY_CONFLICT')
                goal_id = old['goal_id']
            else:
                goal_id = uid()
                policy = self.broker.issue(goal_id, data['action'], data['privacy_class'], token=token)
                priority = score(data['metrics']) if policy['allowed'] else None
                status = 'approved' if policy['allowed'] else 'blocked'
                explanation = {'formula': 'weighted-v1', 'weights': WEIGHTS, 'inputs': data['metrics']} if policy['allowed'] else {'gate': policy['code']}
                timestamp = now()
                db.execute('INSERT INTO goal(id,owner_id,action,privacy_class,status,metrics,computed_priority,priority_explanation,decision_code,assertion,grant_token,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)', (goal_id, 'owner', data['action'], data['privacy_class'], status, json.dumps(data['metrics']), priority, json.dumps(explanation), policy['code'], json.dumps(data['assertion']), policy.get('grant', ''), timestamp, timestamp))
                db.execute('INSERT INTO goal_execution(id,goal_id,idempotency_key,request_hash,execution_state) VALUES(?,?,?,?,?)', (uid(), goal_id, idempotency_key, request_hash, 'queued' if policy['allowed'] else 'blocked'))
                self.event(db, goal_id, 'goal.created', 'OWNER_REQUEST')
                self.event(db, goal_id, 'policy.allowed' if policy['allowed'] else 'policy.denied', policy['code'])
                if priority is not None:
                    self.event(db, goal_id, 'goal.scored', 'WEIGHTED_V1')
        return self.get(goal_id)

    def get(self, goal_id):
        with self.store.connect() as db:
            row = db.execute('SELECT * FROM goal WHERE id=?', (goal_id,)).fetchone()
            if row is None:
                raise KeyError('GOAL_NOT_FOUND')
            out = dict(row)
            out.pop('grant_token')
            for key in ('metrics', 'priority_explanation', 'assertion'):
                out[key] = json.loads(out[key])
            out['evidence'] = [dict(r) for r in db.execute('SELECT * FROM evidence WHERE goal_id=?', (goal_id,))]
            out['execution'] = dict(db.execute('SELECT id,execution_state,attempt,finished_at FROM goal_execution WHERE goal_id=?', (goal_id,)).fetchone())
            out['plan'] = {
                'parent_id': out['parent_id'],
                'depends_on': [r['depends_on_goal_id'] for r in db.execute('SELECT depends_on_goal_id FROM goal_dependency WHERE goal_id=? ORDER BY depends_on_goal_id', (goal_id,))],
            }
            return out

    def set_plan(self, goal_id, depends_on, parent_id, token):
        self.authenticate(token)
        if not isinstance(depends_on, list) or any(not isinstance(item, str) or not item for item in depends_on):
            raise Invalid('PLAN_INVALID')
        if len(depends_on) > 32:
            raise Invalid('PLAN_TOO_LARGE')
        if len(set(depends_on)) != len(depends_on) or parent_id is not None and (not isinstance(parent_id, str) or not parent_id):
            raise Invalid('PLAN_INVALID')
        if goal_id in depends_on or parent_id == goal_id:
            raise Invalid('PLAN_INVALID')
        try:
            with self.store.transaction() as db:
                goal = db.execute('SELECT status FROM goal WHERE id=?', (goal_id,)).fetchone()
                if goal is None:
                    raise KeyError('GOAL_NOT_FOUND')
                if goal['status'] != 'approved':
                    raise Invalid('GOAL_NOT_PLANNABLE')
                references = set(depends_on)
                if parent_id is not None:
                    references.add(parent_id)
                if references:
                    placeholders = ','.join('?' for _ in references)
                    rows = db.execute(f'SELECT id,status,computed_priority FROM goal WHERE id IN ({placeholders})', tuple(references)).fetchall()
                    existing = {r['id'] for r in rows}
                    if existing != references:
                        raise Invalid('PLAN_INVALID')
                    dependencies = {r['id']: r for r in rows if r['id'] in depends_on}
                    if any(row['status'] != 'completed' and (row['status'] != 'approved' or row['computed_priority'] is None or row['computed_priority'] <= 0) for row in dependencies.values()):
                        raise Invalid('DEPENDENCY_NOT_RUNNABLE')
                db.execute('DELETE FROM goal_dependency WHERE goal_id=?', (goal_id,))
                db.execute('UPDATE goal SET parent_id=?,updated_at=?,version=version+1 WHERE id=?', (parent_id, now(), goal_id))
                for dependency_id in sorted(depends_on):
                    db.execute('INSERT INTO goal_dependency(goal_id,depends_on_goal_id,dependency_type,created_at) VALUES(?,?,?,?)', (goal_id, dependency_id, 'blocks', now()))
                self.event(db, goal_id, 'goal.plan.updated', 'OWNER_PLAN')
        except sqlite3.IntegrityError as exc:
            raise Invalid('PLAN_INVALID') from exc
        return {'parent_id': parent_id, 'depends_on': sorted(depends_on)}

    def list_goals(self):
        with self.store.connect() as db:
            return [dict(r) for r in db.execute('SELECT id,status,action,computed_priority,decision_code FROM goal ORDER BY created_at,id LIMIT 100')]

    def events(self, goal_id):
        with self.store.connect() as db:
            return list(reversed([dict(r) for r in db.execute('SELECT * FROM goal_event WHERE goal_id=? ORDER BY rowid DESC LIMIT 100', (goal_id,))]))

    def delivery_counts(self):
        with self.store.connect() as db:
            return tuple(db.execute('SELECT (SELECT count(*) FROM event_outbox WHERE delivered_at IS NOT NULL),(SELECT count(*) FROM event_inbox)').fetchone())

    def tick(self):
        with self.store.transaction() as db:
            candidates = db.execute('SELECT g.* FROM goal g JOIN goal_execution e ON e.goal_id=g.id WHERE e.execution_state="queued" AND g.status="approved" AND g.computed_priority>0 AND NOT EXISTS (SELECT 1 FROM goal_dependency d JOIN goal prerequisite ON prerequisite.id=d.depends_on_goal_id WHERE d.goal_id=g.id AND prerequisite.status!="completed") ORDER BY g.computed_priority DESC,g.created_at,g.id').fetchall()
            for goal in candidates:
                policy = self.broker.evaluate(goal['id'], goal['action'], goal['grant_token'], goal['privacy_class'])
                if not policy['allowed']:
                    if policy['code'] == 'GRANT_INVALID_OR_EXPIRED':
                        db.execute('UPDATE goal SET status="blocked",decision_code="GRANT_RENEWAL_REQUIRED",updated_at=?,version=version+1 WHERE id=?', (now(), goal['id']))
                        db.execute('UPDATE goal_execution SET execution_state="blocked",finished_at=? WHERE goal_id=?', (now(), goal['id']))
                        self.event(db, goal['id'], 'goal.authorization.expired', 'GRANT_RENEWAL_REQUIRED')
                        continue
                    previous = db.execute('SELECT event_type,decision_code FROM goal_event WHERE goal_id=? ORDER BY rowid DESC LIMIT 1', (goal['id'],)).fetchone()
                    if previous is None or (previous['event_type'], previous['decision_code']) != ('policy.denied', policy['code']):
                        self.event(db, goal['id'], 'policy.denied', policy['code'])
                    continue
                goal_id = goal['id']
                self.event(db, goal_id, 'goal.started', 'POLICY_RECHECKED')
                db.execute('UPDATE goal_execution SET execution_state="running",attempt=attempt+1 WHERE goal_id=?', (goal_id,))
                result = execute(goal['action'])
                db.execute('UPDATE goal_execution SET execution_state="verifying" WHERE goal_id=?', (goal_id,))
                passed = verify(result)
                db.execute('INSERT INTO strategy_trial VALUES(?,?,?,?,?,?,?)', (uid(), goal_id, 1, 'Fixed public fixture has expected digest and size', json.dumps(result), 'passed' if passed else 'failed', 'exact_public_fixture_v1'))
                if passed:
                    db.execute('INSERT INTO evidence VALUES(?,?,?,?,?,?)', (uid(), goal_id, result['sha256'], result['byte_count'], 'local_public_fixture_v1', now()))
                db.execute('UPDATE goal SET status=?,updated_at=?,version=version+1 WHERE id=?', ('completed' if passed else 'failed', now(), goal_id))
                db.execute('UPDATE goal_execution SET execution_state=?,finished_at=? WHERE goal_id=?', ('succeeded' if passed else 'failed', now(), goal_id))
                eid = self.event(db, goal_id, 'goal.completed' if passed else 'goal.failed', 'ARTIFACT_VERIFIED' if passed else 'ARTIFACT_MISMATCH')
                # One real local consumer, transactionally idempotent receipt. No
                # claim of distributed bus delivery or external exactly-once effects.
                db.execute('INSERT INTO event_inbox VALUES(?,?)', (eid, now()))
                db.execute('UPDATE event_outbox SET delivered_at=? WHERE event_id=?', (now(), eid))
                return {'outcome': 'succeeded' if passed else 'failed', 'goal_id': goal_id}
        return {'outcome': 'idle_success', 'reason': 'NO_AUTHORIZED_USEFUL_WORK'}
