#!/usr/bin/env python3
"""Native local client: token read from private file, never command arguments."""
import argparse
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
from luna.policy import UnixConnection
from luna.store import read_key

parser = argparse.ArgumentParser()
parser.add_argument('method', choices=['GET', 'POST'])
parser.add_argument('path')
parser.add_argument('--body', help='non-sensitive JSON only')
parser.add_argument('--idempotency-key')
parser.add_argument('--socket')
parser.add_argument('--key-file', default=str(ROOT / '.local/owner.key'))
args = parser.parse_args()
path = args.socket or (ROOT / '.local/api-path').read_text().strip()
headers = {'Authorization': 'Bearer ' + read_key(args.key_file)}
if args.body is not None:
    headers['Content-Type'] = 'application/json'
if args.idempotency_key:
    headers['Idempotency-Key'] = args.idempotency_key
connection = UnixConnection(path)
try:
    connection.request(args.method, args.path, args.body, headers)
    response = connection.getresponse()
    print(json.dumps(json.loads(response.read()), indent=2))
    sys.exit(0 if response.status < 400 else 1)
finally:
    connection.close()
