#!/bin/bash

# Test the real orchestration with fixed privileged paths redirected to harmless
# stand-ins. No host sudo, package transaction, namespace root, or exploit runs.
boundary_tmp=$(mktemp -d)
trap 'rm -rf "$boundary_tmp"' EXIT
export SUDO_TEST_ROOT="$boundary_tmp/lunaos"
export SUDO_TEST_LOG="$boundary_tmp/events"
export SUDO_TEST_CACHE="$boundary_tmp/cache"
export LUNAOS_PATH="$SUDO_TEST_ROOT"
export SUDO_TEST_HOME="$boundary_tmp/home"
# The machine the render reads its region marker from. It starts global; a test
# writes etc/lunaos/region here to model a regional install, never the host's.
export SUDO_TEST_TARGET="$boundary_tmp/target"
export LUNAOS_REGION_FILE="$SUDO_TEST_TARGET/etc/lunaos/region"
mkdir -p "$SUDO_TEST_HOME" "$SUDO_TEST_TARGET/etc"
mkdir -p "$SUDO_TEST_ROOT/bin" "$SUDO_TEST_ROOT/mock" "$SUDO_TEST_ROOT/default/lunaos/sudo-no-update"
: >"$SUDO_TEST_LOG"

copy_boundary_file() {
  python3 - "$ROOT" "$SUDO_TEST_ROOT" "$1" <<'PY'
import sys
from pathlib import Path
source, target, name = map(Path,sys.argv[1:])
p=target/name
p.parent.mkdir(parents=True,exist_ok=True)
s=(source/name).read_text().replace('$HOME', '$SUDO_TEST_HOME')
for command in ['sudo','pkexec','pacman','lunaos-pkg-missing','systemd-inhibit','setpriv','snapper']:
 s=s.replace('/usr/bin/'+command, str(target/'mock'/command))
s=s.replace('PATH=/usr/bin:/usr/sbin:/bin:/sbin', 'PATH="'+str(target/'bin')+':/usr/bin:/usr/sbin:/bin:/sbin"')
if name.name == 'lunaos-apply-pacman':
 render_etc='etc="${4:-/}/etc"'
 assert s.count(render_etc)==1, 'render target lookup moved'
 s=s.replace(render_etc, 'etc="${4:-$SUDO_TEST_TARGET}/etc"')
p.write_text(s)
p.chmod((source/name).stat().st_mode & 0o777)
PY
}

copy_boundary_file bin/lunaos-security-functions
copy_boundary_file bin/lunaos-update-pacman
copy_boundary_file default/lunaos/sudo-no-update/sudo
copy_boundary_file install/helpers/pacman.sh
# Channel changes copy the platform's templates; the copy itself is a stand-in.
cp -r "$ROOT/default/pacman" "$SUDO_TEST_ROOT/default/"

# Channel changes ask for the platform. x86 unless a test says otherwise, so no
# test reads the host's hardware.
cat >"$SUDO_TEST_ROOT/bin/lunaos-hw-platform" <<'STUB'
#!/bin/bash
echo "${SUDO_TEST_PLATFORM:-x86}"
STUB
chmod +x "$SUDO_TEST_ROOT/bin/lunaos-hw-platform"

cat >"$SUDO_TEST_ROOT/mock/sudo" <<'STUB'
#!/bin/bash
set -euo pipefail
printf 'sudo' >>"$SUDO_TEST_LOG"
printf ' %q' "$@" >>"$SUDO_TEST_LOG"
printf '\n' >>"$SUDO_TEST_LOG"
if [[ ${1:-} == "-h" ]]; then
  if [[ ${SUDO_TEST_UNSUPPORTED:-0} == "1" ]]; then
    echo 'usage: sudo [-ABbEHknPS] command'
  else
    echo 'usage: sudo [-ABbEHkNnPS] command'
  fi
  exit 0
fi
if [[ ${1:-} == "-k" || ${1:-} == "-K" ]]; then
  [[ ${SUDO_TEST_REVOKE_FAIL:-0} != "1" ]] || exit 1
  /usr/bin/rm -f "$SUDO_TEST_CACHE"
  exit 0
fi
if [[ ${1:-} == "-n" && ! -e $SUDO_TEST_CACHE ]]; then
  # Non-interactive sudo cannot authenticate without a cached credential.
  exit 1
fi
if [[ ${1:-} == "-N" ]]; then
  shift
else
  touch "$SUDO_TEST_CACHE"
fi
[[ ${SUDO_TEST_SUDO_FAIL:-0} != "1" ]] || exit 1
background=0
while (( $# )); do
  case "$1" in
    -N|-n) shift ;;
    -b) background=1; shift ;;
    -v) exit 0 ;;
    -u|--user) shift 2 ;;
    --) shift; break ;;
    *) break ;;
  esac
done
(( $# )) || exit 0
if (( background )); then
  "$@" &
else
  "$@"
fi
STUB
chmod +x "$SUDO_TEST_ROOT/mock/sudo"
# The updater's own phases call a bare sudo from its fixed PATH. Resolve it to
# the stand-in so no test can ever reach the host's sudo.
ln -s ../mock/sudo "$SUDO_TEST_ROOT/bin/sudo"

cat >"$SUDO_TEST_ROOT/bin/test-step" <<'STUB'
#!/bin/bash
set -euo pipefail
step=${0##*/}
printf 'step:%s %s\n' "$step" "$*" >>"$SUDO_TEST_LOG"
if [[ $step == "systemd-run" ]]; then
  # lunaos-update-pacman registers the transaction as a PID 1 scope on booted
  # hosts. Run the wrapped command in place so the pacman step still executes.
  while (( $# )) && [[ $1 == -* ]]; do shift; done
  exec "$@"
fi
# lunaos update shares one authorization with its post-update hook and mise.
# Standalone hooks, such as the pre-refresh one, and AUR builds run cold.
if [[ $step == "lunaos-hook" && ${1:-} == "post-update" ]] || [[ $step == "lunaos-update-mise" ]]; then
  [[ -e $SUDO_TEST_CACHE ]] || exit 94
elif [[ $step == "lunaos-hook" || $step == "yay" ]]; then
  [[ ! -e $SUDO_TEST_CACHE ]] || exit 91
fi
if [[ -n ${SUDO_TEST_REMOVE_WRAPPER_STEP:-} && "$step $*" == $SUDO_TEST_REMOVE_WRAPPER_STEP ]]; then
  # Model a package transaction replacing the running tree with a release
  # that predates the wrapper.
  /usr/bin/rm -f "$LUNAOS_PATH/default/lunaos/sudo-no-update/sudo"
fi
if [[ ${SUDO_TEST_FAIL_STEP:-} == "$step" ]]; then
  # Model a misbehaving child leaving state behind, then failing. Cleanup must
  # still revoke it. This never invokes real sudo or exercises a privilege flaw.
  touch "$SUDO_TEST_CACHE"
  exit 17
fi
if [[ ${SUDO_TEST_SIGNAL_STEP:-} == "$step" ]]; then
  touch "$SUDO_TEST_CACHE"
  kill -TERM "$PPID"
  exit 0
fi
case "$step" in
  lunaos-update-system-pkgs|lunaos-update-keyring|lunaos-snapshot)
    sudo /usr/bin/true
    ;;
  pacman)
    # A query answers with the packages a test models as installed, one per line.
    if [[ ${1:-} == "-Q" && -n ${SUDO_TEST_INSTALLED:-} ]]; then
      printf '%s\n' "$SUDO_TEST_INSTALLED"
    fi
    ;;
  yay)
    [[ $* == *"--sudo $LUNAOS_PATH/default/lunaos/sudo-no-update/sudo"* ]] || exit 92
    [[ $* == *"--sudoloop=false"* ]] || exit 93
    [[ $(command -v sudo) == "$LUNAOS_PATH/default/lunaos/sudo-no-update/sudo" ]] || exit 95
    ;;
esac
STUB
chmod +x "$SUDO_TEST_ROOT/bin/test-step"
for step in lunaos-update-lock lunaos-update-requires-free-space lunaos-update-confirm lunaos-update-pkg-prune lunaos-snapshot lunaos-update-stay-awake lunaos-update-dev lunaos-update-boot lunaos-update-keyring lunaos-update-system-pkgs lunaos-migrate lunaos-hook lunaos-update-aur-pkgs lunaos-update-mise lunaos-update-orphan-pkgs lunaos-update-analyze-logs lunaos-shell lunaos-update-restart lunaos-pkg-aur-accessible lunaos-notification-dismiss pacman systemd-run cp yay; do
  ln -s test-step "$SUDO_TEST_ROOT/bin/$step"
done
ln -s ../bin/test-step "$SUDO_TEST_ROOT/mock/pacman"

reset_boundary() {
  : >"$SUDO_TEST_LOG"
  /usr/bin/rm -f "$SUDO_TEST_CACHE"
  unset SUDO_TEST_FAIL_STEP SUDO_TEST_SIGNAL_STEP SUDO_TEST_SUDO_FAIL SUDO_TEST_REVOKE_FAIL SUDO_TEST_UNSUPPORTED SUDO_TEST_REMOVE_WRAPPER_STEP SUDO_TEST_PLATFORM SUDO_TEST_INSTALLED
}
assert_boundary_cold() {
  [[ ! -e $SUDO_TEST_CACHE ]] || fail "$1 left cached authorization"
  [[ $(tail -1 "$SUDO_TEST_LOG") == "sudo -k" ]] || fail "$1 did not revoke at exit" "$(<"$SUDO_TEST_LOG")"
}
