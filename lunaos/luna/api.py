"""Owner-authenticated HTTP/JSON served exclusively over a private Unix socket."""
from http.server import BaseHTTPRequestHandler
import json
import os
from pathlib import Path
import socketserver

from .engine import Invalid, Unauthorized
from .store import private_path


class Server(socketserver.UnixStreamServer):
    def __init__(self, path, engine):
        private_path(Path(path).parent, directory=True)
        if os.path.lexists(path):
            raise ValueError('refusing to replace existing socket path')
        self.engine = engine
        super().__init__(path, Handler)
        os.chmod(path, 0o600)


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass  # Never log credentials, request bodies or source assertions.

    def _reply(self, status, value):
        raw = json.dumps(value, allow_nan=False).encode()
        self.send_response(status)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Content-Length', str(len(raw)))
        self.send_header('Cache-Control', 'no-store')
        self.end_headers()
        self.wfile.write(raw)

    def do_GET(self):
        self._handle('GET')

    def do_POST(self):
        self._handle('POST')

    def _handle(self, method):
        self.connection.settimeout(3)
        engine = self.server.engine
        try:
            auth = self.headers.get('Authorization', '')
            engine.authenticate(auth[7:] if auth.startswith('Bearer ') else '')
            if self.headers.get('Origin') is not None:
                raise Unauthorized('BROWSER_ORIGIN_NOT_SUPPORTED')
            body = None
            if method == 'POST':
                if self.headers.get('Content-Type') != 'application/json' or self.headers.get('Transfer-Encoding'):
                    raise Invalid('JSON_CONTENT_TYPE_REQUIRED')
                length = int(self.headers.get('Content-Length', '0'))
                if not 0 < length <= 8192:
                    raise Invalid('BODY_SIZE_INVALID')
                body = json.loads(self.rfile.read(length))
            if method == 'POST' and self.path == '/v1/goals':
                self._reply(201, engine.create(body, auth[7:], self.headers.get('Idempotency-Key')))
            elif method == 'POST' and self.path == '/v1/scheduler/tick':
                if body != {}:
                    raise Invalid('EMPTY_OBJECT_REQUIRED')
                self._reply(200, engine.tick())
            elif method == 'GET' and self.path == '/v1/goals':
                self._reply(200, {'goals': engine.list_goals()})
            elif method == 'GET' and self.path == '/v1/system/capabilities':
                self._reply(200, {'local_goal_slice': 'implemented', 'linux': 'blocked_not_authorized', 'gpu': 'not_implemented', 'sensors': 'not_implemented', 'private_vault': 'not_implemented', 'payments': 'denied_budget_zero', 'remote': 'disabled'})
            elif method == 'GET' and self.path.startswith('/v1/goals/'):
                parts = self.path.split('/')
                if len(parts) == 4:
                    self._reply(200, engine.get(parts[3]))
                elif len(parts) == 5 and parts[4] == 'decisions':
                    engine.get(parts[3])
                    self._reply(200, {'events': engine.events(parts[3])})
                else:
                    self._reply(404, {'code': 'NOT_FOUND'})
            else:
                self._reply(404, {'code': 'NOT_FOUND'})
        except Unauthorized as exc:
            self._reply(401, {'code': str(exc)})
        except Invalid as exc:
            self._reply(400, {'code': str(exc)})
        except (ValueError, TypeError):
            self._reply(400, {'code': 'INVALID_REQUEST'})
        except KeyError:
            self._reply(404, {'code': 'NOT_FOUND'})
        except Exception:
            self._reply(503, {'code': 'LOCAL_SERVICE_UNAVAILABLE'})
