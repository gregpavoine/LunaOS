#!/bin/bash
# Fixture tests for rebrand.sh: each test builds a throwaway git repo.
set -euo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REBRAND="$HERE/rebrand.sh"
ROOT=$(mktemp -d "${TMPDIR:-/tmp}/rebrand-test.XXXXXX")
trap 'rm -rf "$ROOT"' EXIT

passed=0
failed=0

assert_eq() {
  local label=$1 expected=$2 actual=$3
  if [[ $expected == "$actual" ]]; then
    passed=$((passed + 1))
  else
    failed=$((failed + 1))
    printf 'FAIL: %s\n  expected: %s\n  actual:   %s\n' "$label" "$expected" "$actual"
  fi
}

# new_repo <name>: prints the path of an empty git repo
new_repo() {
  local repo="$ROOT/$1"
  mkdir -p "$repo"
  git -C "$repo" init -q
  git -C "$repo" config user.name test
  git -C "$repo" config user.email test@example.com
  echo "$repo"
}

# write_file <repo> <path> <content>
write_file() {
  mkdir -p "$(dirname "$1/$2")"
  printf '%s\n' "$3" >"$1/$2"
}

commit_all() {
  git -C "$1" add -A
  git -C "$1" commit -q -m "${2:-fixture}"
}

# run_rebrand <args...>: sets rc and out
run_rebrand() {
  rc=0
  out=$(bash "$REBRAND" "$@" 2>&1) || rc=$?
}

test_renames_paths_and_contents() {
  local repo
  repo=$(new_repo renames)
  write_file "$repo" bin/omarchy-foo "OMARCHY_PATH Omarchy omarchy"
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "renames: exit code" 0 "$rc"
  assert_eq "renames: old path gone" "no" "$([[ -e $repo/bin/omarchy-foo ]] && echo yes || echo no)"
  assert_eq "renames: contents" "LUNAOS_PATH LunaOS lunaos" "$(cat "$repo/bin/lunaos-foo")"
}

test_nested_path() {
  local repo
  repo=$(new_repo nested)
  write_file "$repo" default/omarchy/omarchy-menu.jsonc "{}"
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "nested: exit code" 0 "$rc"
  assert_eq "nested: new path" "yes" "$([[ -f $repo/default/lunaos/lunaos-menu.jsonc ]] && echo yes || echo no)"
  assert_eq "nested: old dir gone" "no" "$([[ -e $repo/default/omarchy ]] && echo yes || echo no)"
}

test_longest_exception_wins() {
  local repo
  repo=$(new_repo longest)
  write_file "$repo" notes.txt "linux-omarchy-headers linux-omarchy omarchy-nvim omarchy-update"
  write_file "$repo" pkgs.txt "omarchy-mac omarchy-mac-boot omarchy-mac-pkgs omarchy-keyring"
  write_file "$repo" bin/omarchy-update "x"
  write_file "$repo" etc/omarchy-mac-encrypt.conf "x"
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "longest: exit code" 0 "$rc"
  assert_eq "longest: contents" "linux-omarchy-headers linux-omarchy omarchy-nvim lunaos-update" "$(cat "$repo/notes.txt")"
  assert_eq "longest: mac packages" "omarchy-mac omarchy-mac-boot omarchy-mac-pkgs omarchy-keyring" "$(cat "$repo/pkgs.txt")"
  assert_eq "longest: path renamed" "yes" "$([[ -f $repo/bin/lunaos-update ]] && echo yes || echo no)"
  assert_eq "longest: path exception kept" "yes" "$([[ -f $repo/etc/omarchy-mac-encrypt.conf ]] && echo yes || echo no)"
}

test_update_repo_rewritten() {
  local repo
  repo=$(new_repo repourl)
  write_file "$repo" url.txt "https://github.com/omacom/omarchy.git https://github.com/omacom/omarchy/releases omacom/omarchy-pkgs"
  write_file "$repo" links.txt "https://github.com/omacom/omarchy/discussions/1"
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "repourl: exit code" 0 "$rc"
  assert_eq "repourl: contents" "https://github.com/gregpavoine/LunaOS.git https://github.com/gregpavoine/LunaOS/releases omacom/omarchy-pkgs" "$(cat "$repo/url.txt")"
  assert_eq "repourl: upstream link kept" "https://github.com/omacom/omarchy/discussions/1" "$(cat "$repo/links.txt")"
}

test_skips_lunaos_dir() {
  local repo
  repo=$(new_repo skiplunaos)
  write_file "$repo" lunaos/docs/x.md "omarchy"
  write_file "$repo" LICENSE "omarchy"
  write_file "$repo" test/shell.d/fixtures/omarchy-mac-boot-mkinitcpio/90-omarchy-mac.conf "_omarchy_asahi_hooks"
  write_file "$repo" bin/omarchy-font-list "grep -v -i -E 'emoji|signwriting|omarchy'"
  write_file "$repo" other.txt "omarchy"
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "skip: exit code" 0 "$rc"
  assert_eq "skip: lunaos untouched" "omarchy" "$(cat "$repo/lunaos/docs/x.md")"
  assert_eq "skip: LICENSE untouched" "omarchy" "$(cat "$repo/LICENSE")"
  assert_eq "skip: upstream fixture untouched" "_omarchy_asahi_hooks" "$(cat "$repo/test/shell.d/fixtures/omarchy-mac-boot-mkinitcpio/90-omarchy-mac.conf")"
  assert_eq "skip: others rewritten" "lunaos" "$(cat "$repo/other.txt")"
  assert_eq "skip: icon font family filter kept" "grep -v -i -E 'emoji|signwriting|omarchy'" "$(cat "$repo/bin/lunaos-font-list")"
}

test_binary_content_untouched() {
  local repo before
  repo=$(new_repo binary)
  mkdir -p "$repo/default/fonts/omarchy"
  printf '\0omarchy\0' >"$repo/default/fonts/omarchy/omarchy.ttf"
  before=$(cksum <"$repo/default/fonts/omarchy/omarchy.ttf")
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "binary: exit code" 0 "$rc"
  assert_eq "binary: path renamed" "yes" "$([[ -f $repo/default/fonts/lunaos/lunaos.ttf ]] && echo yes || echo no)"
  assert_eq "binary: bytes identical" "$before" "$(cksum <"$repo/default/fonts/lunaos/lunaos.ttf")"
}

test_unknown_case_fails() {
  local repo
  repo=$(new_repo unknowncase)
  write_file "$repo" bin/omarchy-foo "omarchy"
  write_file "$repo" weird.txt "OmArchy"
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "unknown case: exit code" 3 "$rc"
  assert_eq "unknown case: tree unchanged" "" "$(git -C "$repo" status --porcelain)"
  assert_eq "unknown case: path kept" "yes" "$([[ -f $repo/bin/omarchy-foo ]] && echo yes || echo no)"
}

test_idempotent() {
  local repo
  repo=$(new_repo idempotent)
  write_file "$repo" default/omarchy/omarchy-menu.jsonc "Omarchy omarchy OMARCHY"
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "idempotent: first exit code" 0 "$rc"
  commit_all "$repo" first
  run_rebrand "$repo"
  assert_eq "idempotent: second exit code" 0 "$rc"
  assert_eq "idempotent: tree clean" "" "$(git -C "$repo" status --porcelain)"
}

test_check_mode() {
  local repo
  repo=$(new_repo check)
  write_file "$repo" notes.txt "Omarchy omarchy-nvim"
  commit_all "$repo"
  run_rebrand "$repo"
  commit_all "$repo" first
  run_rebrand --check "$repo"
  assert_eq "check: clean exit code" 0 "$rc"
  printf 'still omarchy here\n' >>"$repo/notes.txt"
  run_rebrand --check "$repo"
  assert_eq "check: dirty exit code" 1 "$rc"
  assert_eq "check: line listed" "yes" "$([[ $out == *notes.txt:2* ]] && echo yes || echo no)"
}

test_dirty_tree_refused() {
  local repo
  repo=$(new_repo dirty)
  write_file "$repo" a.txt "omarchy"
  commit_all "$repo"
  printf 'x\n' >>"$repo/a.txt"
  run_rebrand "$repo"
  assert_eq "dirty: exit code" 2 "$rc"
  run_rebrand "$ROOT"
  assert_eq "not a git tree: exit code" 2 "$rc"
}

test_internal_failure_is_fatal() {
  local repo copy
  repo=$(new_repo failure)
  write_file "$repo" a.txt "omarchy"
  commit_all "$repo"
  # A copy of the script with no exceptions.txt next to it: the real file is never moved
  copy=$(mktemp -d "${TMPDIR:-/tmp}/rebrand-noexc.XXXXXX")
  cp "$HERE/rebrand.sh" "$copy/rebrand.sh"
  rc=0
  "$copy/rebrand.sh" "$repo" >/dev/null 2>&1 || rc=$?
  assert_eq "failure: rewrite exit code" 5 "$rc"
  rc=0
  "$copy/rebrand.sh" --check "$repo" >/dev/null 2>&1 || rc=$?
  assert_eq "failure: --check exit code" 5 "$rc"
  rm -rf "$copy"
  assert_eq "failure: tree unchanged" "" "$(git -C "$repo" status --porcelain)"
}

test_check_restores_exception_in_listing() {
  local repo
  repo=$(new_repo restore)
  write_file "$repo" a.txt "omarchy-nvim and omarchy"
  commit_all "$repo"
  run_rebrand --check "$repo"
  assert_eq "restore: exit code" 1 "$rc"
  assert_eq "restore: line unmasked" "yes" "$([[ $out == *'omarchy-nvim and omarchy'* ]] && echo yes || echo no)"
}

test_symlink_and_target_collision() {
  local repo
  repo=$(new_repo symlink)
  write_file "$repo" real.txt "omarchy"
  ln -s real.txt "$repo/link.txt"
  write_file "$repo" "x/omarchy-a" "x"
  write_file "$repo" "x/lunaos-a" "y"
  commit_all "$repo"
  run_rebrand "$repo"
  assert_eq "existing target: exit code" 4 "$rc"
  git -C "$repo" rm -q -f "x/lunaos-a"
  git -C "$repo" commit -q -m drop
  run_rebrand "$repo"
  assert_eq "symlink: exit code" 0 "$rc"
  assert_eq "symlink: still a link" "yes" "$([[ -L $repo/link.txt ]] && echo yes || echo no)"
  assert_eq "symlink: target rewritten once" "lunaos" "$(cat "$repo/real.txt")"
}

for t in test_renames_paths_and_contents test_nested_path test_longest_exception_wins \
  test_update_repo_rewritten test_skips_lunaos_dir test_binary_content_untouched \
  test_unknown_case_fails test_idempotent test_check_mode \
  test_dirty_tree_refused test_internal_failure_is_fatal test_check_restores_exception_in_listing \
  test_symlink_and_target_collision; do
  echo "== $t"
  "$t"
done

printf '%d passed, %d failed\n' "$passed" "$failed"
if (( failed > 0 )); then
  exit 1
fi
