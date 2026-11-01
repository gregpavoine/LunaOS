# LunaOS package signing key

`lunaos-packages.asc` is the public half of the key that signs LunaOS packages and repository databases.

- Fingerprint: `48C0A74703A61B358247549A5EA1A2D955727D01`
- Type: Ed25519, signing only, created 2026-11-01, expires 2028-10-31
- Private half: the `LUNAOS_PACKAGE_SIGNING_KEY` repository secret, plus an offline backup kept by the owner; never in Git

`lunaos/packaging/control-plane/build-signed-repo.sh` refuses to sign with any other key, and checks that pacman accepts this key's repository and refuses an unsigned one or one signed by another key.

To rotate: generate a new key, replace this file and the fingerprint above, update the secret, and ship the new public key to installed systems before packages signed only with it.
