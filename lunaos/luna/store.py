"""Versioned durable local repository, public-fixture metadata only."""
from contextlib import contextmanager
from pathlib import Path
import os
import sqlite3
import stat

MIGRATION = Path(__file__).resolve().parents[1] / 'migrations/core/001_bootstrap.sql'


def private_path(path, directory=False):
    info = Path(path).lstat()
    if stat.S_ISLNK(info.st_mode) or (directory and not stat.S_ISDIR(info.st_mode)) or (not directory and not stat.S_ISREG(info.st_mode)) or info.st_mode & 0o077 or info.st_uid != os.getuid():
        raise ValueError('owner-only regular file/private directory required')


def read_key(path):
    private_path(path)
    token = Path(path).read_text().strip()
    if len(token) != 64 or any(c not in '0123456789abcdef' for c in token):
        raise ValueError('owner key must be 64 lowercase hex characters')
    return token


class Store:
    def __init__(self, path):
        self.path = Path(path).absolute()
        private_path(self.path.parent, directory=True)
        if not self.path.exists():
            fd = os.open(self.path, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
            os.close(fd)
        private_path(self.path)
        with self.connect() as db:
            db.executescript(MIGRATION.read_text())

    @contextmanager
    def connect(self):
        db = sqlite3.connect(self.path, timeout=10, isolation_level=None)
        db.row_factory = sqlite3.Row
        db.execute('PRAGMA foreign_keys=ON')
        db.execute('PRAGMA synchronous=FULL')
        db.execute('PRAGMA busy_timeout=10000')
        try:
            yield db
        finally:
            db.close()

    @contextmanager
    def transaction(self):
        with self.connect() as db:
            db.execute('BEGIN IMMEDIATE')
            try:
                yield db
                db.commit()
            except BaseException:
                db.rollback()
                raise
