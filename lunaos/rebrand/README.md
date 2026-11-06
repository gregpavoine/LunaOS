# Rebrand

`rebrand.sh` is the replayable Omarchy to LunaOS rename; `exceptions.txt` lists the strings that keep the upstream name on purpose, and `rebrand-test.sh` tests the script. See the header of `rebrand.sh` for usage and exit codes.

## Upstream sync

1. Advance the `upstream` branch to the new upstream release.
2. Check it out as a copy (a worktree), run `lunaos/rebrand/rebrand.sh <copy>` on it, and commit the result as the generated commit. The generated commit's parent is LunaOS history (707943c8), not `upstream`, and its content is pure script output.
3. Merge it into the LunaOS branch and resolve conflicts in the files LunaOS edited after generation: `README.md`, `.github/SECURITY.md`, the agents' contributing and reporting docs, `install/lunaos-base.packages`, `install/user/theme.sh`, `config/lunaos/shell.json`, `default/hypr/**`, `default/lunaos/launcher.hides`, the manual, and `default/i18n/en/keybindings.json`.
4. Run `lunaos/rebrand/rebrand.sh --check .`; it must exit 0. A new upstream string that must keep the name needs a narrow literal entry in `exceptions.txt`, which never occurs on `upstream` in LunaOS-owned text.
5. Re-check the Series 26 removals. New upstream references to the removed applications arrive silently, and only the package list is guarded by a test.

Symlink targets are not rewritten by the script; check any new or changed symlink by hand.
