# Intimacy and private memory checkpoint

This document describes the currently implemented Go slice. It is not a claim
that the Mont Blanc Intimacy Core is complete.

## Implemented

- `POST /v1/intimacy/stop` persists a fail-closed stop state in the Core database.
- `GET /v1/intimacy/relationship/state` exposes only stop, privacy-mode and backend
  availability metadata. New or reopened Core starts stopped with privacy `unknown`.
- `POST /v1/intimacy/conversation` is denied while STOP is active or privacy is
  unknown. If those gates later pass, it still returns `LOCAL_MODEL_UNAVAILABLE`
  because no local inference backend is configured.
- The optional separate `services/intimacy` SQLite vault encrypts memory payloads
  with AES-256-GCM. Associated data binds ciphertext to the memory ID, owner and
  provenance type. The database parent directory must already be private (`0700`);
  the DB must be owner-only (`0600`).
- Memory read/write requires a random consent ID with the corresponding scope and
  an unexpired, unrevoked consent. API consent creation requires owner confirmation
  and adult self-attestation. This self-attestation is not identity or age
  verification. Only the local owner subject is supported; other subjects are
  rejected.
- Memory deletion remains available after consent revocation. Audit rows record
  consent grant/revoke and memory deletion using non-sensitive event codes.
- Explicit owner preferences are persisted separately from memories and encrypted
  with AES-256-GCM in the same private vault. Only `declared` provenance is
  accepted; no inferred or model-generated preference is fabricated. Reads and
  writes require distinct `preferences.read` and `preferences.write` scopes, and
  write/correction responses expose metadata without the private value.
- Preference correction replaces the declared value, preserves first-seen time,
  increments evidence count, and records disagreement as negative feedback.
  Owner deletion remains available after consent revocation. Audit events contain
  only fixed event codes and timestamps.
- Discovery policy stores source, creator, category, model-family and tag
  allowlists encrypted in the private vault. Policy read, update and candidate
  evaluation have separate expiring consent scopes. A candidate is allowed only
  when every required field and every supplied tag matches the configured lists;
  the default policy is disabled. Updating the policy requires owner confirmation.
  Evaluation handles caller-supplied metadata only and stores neither candidates
  nor media.
- The vault can export a consistent SQLite snapshot into an authenticated,
  chunk-encrypted backup using a separate 32-byte backup key. Restore decrypts to
  a private temporary file, checks SQLite integrity and every encrypted payload,
  then publishes only to a new destination without overwriting an existing file.
  Backup and restore are library operations; they do not manage keys or coordinate
  service shutdown.

## Optional local configuration

The Core starts with Intimacy storage disabled unless both variables are present:

```text
LUNA_INTIMACY_DATABASE=/private/path/intimacy/vault.sqlite
LUNA_INTIMACY_KEY_FILE=/private/path/intimacy/vault.key
```

The parent directory must exist and be owned by the service account with no group
or other permissions. The key file must be a regular, non-symlink, owner-only file
and contain 32 raw bytes or 64 hexadecimal characters. `luna-vault keygen` creates
random 32-byte keys in owner-only files without displaying them or replacing an
existing path. `luna-vault backup` and `luna-vault restore` wrap the encrypted
backup/restore operations and require separate owner-only vault and backup key
files. Restore requires a new destination and refuses to overwrite files. The
library also exposes an owner-confirmed atomic `RotateKey` operation, but the CLI
does not rotate or replace key files. The software does not store keys in an OS
key store, coordinate rotation with a running Core process, or recover a lost
vault or backup key. Retain both keys securely. Stop all vault users and verify
an encrypted backup before rotating; replace the configured vault key only after
successful rotation and verification. The command has not been validated on
Linux. Do not use the general Core database or ordinary logs for private payloads.

Example commands (all key paths must refer to owner-only regular files; the
destination directory must already be owner-private):

```sh
luna-vault keygen --destination /private/intimacy/vault.key
luna-vault keygen --destination /private/intimacy/backup.key

luna-vault backup --database /private/intimacy/vault.sqlite \
  --vault-key-file /private/intimacy/vault.key \
  --backup-key-file /private/intimacy/backup.key \
  --destination /private/backups/intimacy.lunabak

luna-vault restore --backup-file /private/backups/intimacy.lunabak \
  --backup-key-file /private/intimacy/backup.key \
  --vault-key-file /private/intimacy/vault.key \
  --destination /private/intimacy/restored.sqlite
```

Routes require the local owner bearer token. `Origin` browser requests are rejected.
Consent scopes are currently `memory.read`, `memory.write`, `preferences.read`,
and `preferences.write`, plus `discovery.policy.read`,
`discovery.policy.write`, and `discovery.run`; the random consent ID is required
for each operation.
Memory and preference write responses return opaque IDs and safe metadata;
reading plaintext is a separate scoped operation.

## Not implemented

No local conversation model, content-safety classifier, preference learner, UI or
voice stop control, BLIND sensor integration, third-party subject verification,
operator backup/key-management workflow, automatic retention/derivative erasure, adult media
indexing/search/recommendation, source crawler, ComfyUI bridge or GPU scheduling is
present. The allowlist gate does not fetch or inspect media, establish licensing,
verify anyone's age or consent, or detect unknown people. The preference API
stores only choices the owner explicitly declares; it does not discover or infer
user preferences. Do not enable this as a general adult-content service. The
release matrix keeps the full Intimacy Core marked PARTIAL.
