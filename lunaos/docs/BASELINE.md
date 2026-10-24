# Non-destructive baseline — 2026-10-24

## Observed local repository/controller

- Repository `/Users/gregorypavoine/Projects/LunaOS`, existing branch `main`,
  initial source checkpoint `cadb5c2` (Initial LunaOS specifications).
- Required root AGENTS and both named specs read before edits. Also read Series26
  Omarchy contract; supplied duplicated start-pack preserved.
- `python3 --version`: Python 3.11.7; `go version`: go1.20.6 darwin/arm64.
- `/usr/bin/git` route reported unaccepted Xcode license; `/usr/local/bin/git`
  reported `bad CPU type in executable`. No license/system settings changed.
- `/Library/Developer/CommandLineTools/usr/bin/git` operates successfully.
- Local Go builds use `-buildvcs=false` to avoid the incompatible Git found in PATH.
  Actual checkpoint provenance comes from Git, not embedded automatic VCS stamping.
- Development venv generated with `--without-pip`; no third-party runtime download.

## Linux baseline — BLOCKED / NOT EXECUTED

No SSH alias or connection authorization supplied. No remote connection attempted.
No Omarchy/macOS/Linux packages, services, boot, partitions, mounts or drivers changed.
`scripts/linux-baseline.sh` authored and syntax-checked only. Existing cleanup
scripts never executed. Target packages/GPU/mount/network/ComfyUI state UNKNOWN.
