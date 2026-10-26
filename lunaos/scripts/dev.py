#!/usr/bin/env python3
"""Start only the local development broker and API; never connects to Linux."""
import os
import getpass
from pathlib import Path
import secrets
import signal
import subprocess
import sys
import tempfile
import time

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
from luna.store import private_path, read_key


def main():
    os.umask(0o077)
    local = ROOT / '.local'
    local.mkdir(mode=0o700, exist_ok=True)
    private_path(local, directory=True)
    key = local / 'owner.key'
    if not key.exists():
        fd = os.open(key, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(fd, 'w') as stream:
            stream.write(secrets.token_hex(32))
    read_key(key)
    session = Path(tempfile.mkdtemp(prefix='run-', dir=local))
    binary = session / 'policy'
    subprocess.run(['go', 'build', '-buildvcs=false', '-o', str(binary), './services/policy'], cwd=ROOT, check=True)
    broker_socket = session / 'policy.sock'
    api_socket = session / 'api.sock'
    processes = []
    def stop(*_):
        raise KeyboardInterrupt
    signal.signal(signal.SIGTERM, stop)
    try:
        processes.append(subprocess.Popen([str(binary), '--socket', str(broker_socket), '--key-file', str(key), '--client-user', getpass.getuser()]))
        for _ in range(100):
            if broker_socket.exists():
                break
            if processes[0].poll() is not None:
                raise RuntimeError('broker startup failed')
            time.sleep(.02)
        else:
            raise RuntimeError('broker startup timed out')
        processes.append(subprocess.Popen([sys.executable, '-m', 'luna', 'serve', '--db', str(local / 'goals.sqlite'), '--broker-socket', str(broker_socket), '--key-file', str(key), '--api-socket', str(api_socket), '--auto-run'], cwd=ROOT))
        for _ in range(100):
            if api_socket.exists():
                break
            if processes[1].poll() is not None:
                raise RuntimeError('API startup failed')
            time.sleep(.02)
        else:
            raise RuntimeError('API startup timed out')
        (local / 'api-path').write_text(str(api_socket))
        print(f'Local API socket: {api_socket}\nOwner key file: {key}\nOnly public fixture metadata is supported. Ctrl-C stops these local processes.', flush=True)
        while all(p.poll() is None for p in processes):
            time.sleep(.2)
        raise RuntimeError('local service exited; no hardware fallback')
    except KeyboardInterrupt:
        pass
    finally:
        for process in reversed(processes):
            if process.poll() is None:
                process.terminate()
                process.wait(timeout=5)
        # Only this launcher's own private session sockets are removed.
        for path in (api_socket, broker_socket):
            path.unlink(missing_ok=True)


if __name__ == '__main__':
    main()
