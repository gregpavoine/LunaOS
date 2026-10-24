#!/usr/bin/env python3
"""Exercise documented local launcher/client through a full process restart."""
import json
from pathlib import Path
import subprocess
import sys
import time
import uuid

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
from luna.policy import UnixConnection
from luna.store import read_key


def launch():
    marker = ROOT / '.local/api-path'
    previous = marker.read_text() if marker.exists() else None
    process = subprocess.Popen([sys.executable, 'scripts/dev.py'], cwd=ROOT, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    for _ in range(300):
        if process.poll() is not None:
            raise RuntimeError(process.stderr.read())
        if marker.exists() and marker.read_text() != previous:
            path = marker.read_text().strip()
            if Path(path).exists():
                return process, path
        time.sleep(.05)
    process.terminate()
    process.wait(timeout=10)
    raise RuntimeError('local launcher timeout')


def stop(process):
    process.terminate()
    process.wait(timeout=10)
    process.stdout.close()
    process.stderr.close()
    if process.returncode != 0:
        raise RuntimeError('launcher did not stop cleanly')


def request(path, method, route, body=None, key=None):
    conn = UnixConnection(path)
    headers = {'Authorization': 'Bearer ' + read_key(ROOT / '.local/owner.key')}
    if body is not None:
        headers['Content-Type'] = 'application/json'
    if key:
        headers['Idempotency-Key'] = key
    try:
        conn.request(method, route, json.dumps(body) if body is not None else None, headers)
        response = conn.getresponse()
        value = json.loads(response.read())
        assert response.status < 400, value
        return value
    finally:
        conn.close()


def main():
    key = 'smoke-' + str(uuid.uuid4())
    process, path = launch()
    try:
        goal = request(path, 'POST', '/v1/goals', {'action': 'local.checksum'}, key)
    finally:
        stop(process)
    process, path = launch()
    try:
        duplicate = request(path, 'POST', '/v1/goals', {'action': 'local.checksum'}, key)
        assert duplicate['id'] == goal['id']
        for _ in range(100):
            final = request(path, 'GET', '/v1/goals/' + goal['id'])
            if final['status'] == 'completed':
                break
            time.sleep(.05)
        assert final['status'] == 'completed', final
        assert len(final['evidence']) == 1
        print(json.dumps({'launcher_restart': 'PASS', 'persisted_goal': goal['id'], 'status': final['status'], 'evidence_count': len(final['evidence']), 'sha256': final['evidence'][0]['sha256'], 'hardware': 'NOT_TESTED'}))
    finally:
        stop(process)


if __name__ == '__main__':
    main()
