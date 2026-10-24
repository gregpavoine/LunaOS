"""HTTP over private Unix socket; no TCP fallback or local policy bypass."""
import http.client
import json
import socket


class UnixConnection(http.client.HTTPConnection):
    def __init__(self, path):
        super().__init__('localhost', timeout=3)
        self.path = path

    def connect(self):
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.settimeout(self.timeout)
        self.sock.connect(self.path)


class Broker:
    def __init__(self, path, token):
        self.path, self.token = path, token

    def _request(self, path, goal_id, action, grant='', privacy='general', token=None):
        connection = UnixConnection(self.path)
        try:
            connection.request('POST', path, json.dumps({'goal_id': goal_id, 'action': action, 'privacy_class': privacy, 'grant': grant}), {'Authorization': 'Bearer ' + (self.token if token is None else token), 'Content-Type': 'application/json'})
            response = connection.getresponse()
            result = json.loads(response.read(16384))
            if response.status not in (200, 401) or type(result.get('allowed')) is not bool:
                raise ValueError('invalid broker response')
            return result
        except (OSError, ValueError, http.client.HTTPException):
            return {'allowed': False, 'code': 'BROKER_UNAVAILABLE'}
        finally:
            connection.close()

    def issue(self, goal_id, action, privacy='general', token=None):
        return self._request('/v1/grants', goal_id, action, privacy=privacy, token=token)

    def evaluate(self, goal_id, action, grant, privacy='general'):
        return self._request('/v1/policy/evaluate', goal_id, action, grant, privacy)
