"""Portable local worker/API entrypoint. No target-host operations."""
import argparse
import json
from pathlib import Path
import threading

from .api import Server
from .engine import Engine
from .policy import Broker
from .store import read_key


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('command', choices=['serve', 'tick'])
    parser.add_argument('--db', required=True)
    parser.add_argument('--broker-socket', required=True)
    parser.add_argument('--key-file', required=True)
    parser.add_argument('--api-socket')
    parser.add_argument('--auto-run', action='store_true', help='run one queued action per second')
    args = parser.parse_args()
    engine = Engine(args.db, Broker(args.broker_socket, read_key(args.key_file)))
    if args.command == 'tick':
        print(json.dumps(engine.tick()))
        return
    if not args.api_socket:
        parser.error('--api-socket required for serve')
    stop = threading.Event()
    def worker():
        while not stop.wait(1):
            try:
                engine.tick()
            except Exception:
                # Fail closed; transaction rolled back. Public operational code only.
                print(json.dumps({'code': 'WORKER_FAILED_WORK_RETAINED'}), flush=True)
    with Server(args.api_socket, engine) as server:
        if args.auto_run:
            threading.Thread(target=worker, daemon=True).start()
        try:
            server.serve_forever(poll_interval=.2)
        finally:
            stop.set()
            Path(args.api_socket).unlink(missing_ok=True)


if __name__ == '__main__':
    main()
