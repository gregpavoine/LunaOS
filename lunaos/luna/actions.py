"""One real bounded pure action. No shell, arbitrary paths or external I/O."""
import hashlib
from typing import Protocol

PUBLIC_FIXTURE = b'LunaOS public checkpoint v1\n'


class LinuxAdapter(Protocol):
    """Future authenticated target adapter. No implementation enabled in Task 1."""
    def baseline(self) -> dict: ...
    def service_status(self, unit: str) -> dict: ...
    def submit_comfyui(self, workflow_id: str, grant_id: str) -> dict: ...


def execute(action):
    if action != 'local.checksum':
        raise ValueError('action not implemented')
    return {'sha256': hashlib.sha256(PUBLIC_FIXTURE).hexdigest(), 'byte_count': len(PUBLIC_FIXTURE)}


def verify(result):
    # Separate evaluator checks actual returned digest and length against fixture.
    expected = hashlib.new('sha256', PUBLIC_FIXTURE).hexdigest()
    return result == {'sha256': expected, 'byte_count': len(PUBLIC_FIXTURE)}
