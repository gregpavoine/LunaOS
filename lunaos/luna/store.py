"""Versioned durable local repository, public-fixture metadata only."""
from contextlib import contextmanager
from pathlib import Path
import os
import sqlite3
import stat

MIGRATIONS = Path(__file__).resolve().parents[1] / 'migrations/core'


def statements(script):
    current = ''
    for line in script.splitlines(keepends=True):
        current += line
        if sqlite3.complete_statement(current):
            statement, current = current.strip(), ''
            if statement:
                yield statement
    if current.strip():
        raise ValueError('incomplete database migration')


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
            try:
                fd = os.open(self.path, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
            except FileExistsError:
                pass
            else:
                os.close(fd)
        private_path(self.path)
        migration_files = sorted(MIGRATIONS.glob('[0-9][0-9][0-9]_*.sql'))
        for migration in migration_files:
            version = int(migration.name.split('_', 1)[0])
            with self.connect() as db:
                db.execute('BEGIN IMMEDIATE')
                try:
                    migration_table = db.execute("SELECT 1 FROM sqlite_master WHERE type='table' AND name='schema_migration'").fetchone()
                    applied = db.execute('SELECT 1 FROM schema_migration WHERE version=?', (version,)).fetchone() if migration_table else None
                    if applied is None:
                        for statement in statements(migration.read_text()):
                            if statement.rstrip(';').strip().upper() not in ('BEGIN IMMEDIATE', 'COMMIT'):
                                db.execute(statement)
                    db.commit()
                except BaseException:
                    db.rollback()
                    raise

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
