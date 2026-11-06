# Install Scripts

Read this before working under `install/` or on the system/user setup commands.

The ISO owns installation orchestration. This repo ships target-side setup
commands and reusable setup leaves:

- `bin/lunaos-apply-system` runs root-owned system setup during ISO finalization.
- `bin/lunaos-apply-hardware` runs idempotent hardware-specific setup and is called by `lunaos-apply-system`.
- `bin/lunaos-finalize-user` runs the per-user runtime finalization (skill symlinks, xdg-user-dirs, mime defaults, `install/user/all.sh`). Shipped user defaults are seeded by `/etc/skel` from `lunaos-settings`, not by this command. `bin/lunaos-reinstall-configs` is the explicit destructive resync of those defaults into an existing user's `$HOME`.
- leaf scripts under `install/` are sourced by `run_logged $LUNAOS_INSTALL/path/to/script.sh` and intentionally do not have shebangs.
- avoid `exit` in sourced setup scripts unless intentionally aborting setup.
- use `$LUNAOS_INSTALL` and `$LUNAOS_PATH` instead of hard-coded LunaOS paths.
- keep root-scoped hardware setup under `install/hardware/` and orchestrate it through `install/hardware/all.sh`.
- in an image build every `run_logged` leaf in `install/hardware/all.sh` is queued and runs on the machine's first boot instead (`bin/lunaos-provision-hardware`, see `docs/file-layout.md`), possibly offline, using packages already in the image, and before any user exists; keep those leaves idempotent and put nothing but `run_logged` lines in `all.sh`.
- gate platform-specific setup on a predicate built on `lunaos-hw-platform` (`lunaos-hw-aarch64-apple`), not on `uname -m`; architecture checks (`lunaos-hw-x86`, `lunaos-hw-aarch64`) are only for ABI, binary availability or a repository's `$arch`. Platform names and where platform code lives are in the Platforms section of `AGENTS.md`.
- keep every per-user setup leaf under `install/user/` (including `install/user/hardware/` and `install/user/first-run/`) so it is clear what must run for each user.
- The base install supplies matching kernel headers before hardware setup. DKMS installers should install their driver packages and assume the headers exist.
- prefer helper commands for package and command checks where available.

Raw `command -v`, `pacman`, and `pacman-key` are acceptable in package-helper
contexts where direct package-manager behavior is the point of the script.
