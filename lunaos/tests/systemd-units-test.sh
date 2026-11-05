#!/bin/bash
# Static hardening assertions for the LUNA systemd units (run anywhere), plus
# systemd-analyze verify inside a disposable Linux container only.

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
units=$root/systemd
failures=0

fail() {
  echo "FAIL: $1" >&2
  failures=$((failures + 1))
}

# section_lines <file> <section>: directive lines of one [Section]
section_lines() {
  awk -v want="[$2]" '
    /^\[/ { current = $0; next }
    current == want && NF && $0 !~ /^[#;]/ { print }
  ' "$1"
}

# expect_exact <unit> <section> <key> <value>: key set exactly once to value
expect_exact() {
  local unit=$1 section=$2 key=$3 value=$4 found
  found=$(section_lines "$units/$unit" "$section" | grep -E "^$key=" || true)
  if [[ $found != "$key=$value" ]]; then
    fail "$unit [$section] expected '$key=$value', found '${found:-<unset>}'"
  fi
}

for unit in luna-core.service luna-policy-broker.service; do
  expect_exact "$unit" Service NoNewPrivileges yes
  expect_exact "$unit" Service ProtectSystem strict
  expect_exact "$unit" Service ProtectHome yes
  expect_exact "$unit" Service PrivateTmp yes
  expect_exact "$unit" Service Restart on-failure
  expect_exact "$unit" Unit StartLimitBurst 5
  expect_exact "$unit" Unit StartLimitIntervalSec 60
done

expect_exact luna-core.service Service User lunaos
expect_exact luna-core.service Service RestrictAddressFamilies "AF_UNIX AF_INET"
expect_exact luna-policy-broker.service Service User luna-policy
expect_exact luna-policy-broker.service Service RestrictAddressFamilies AF_UNIX

expect_exact luna-owner-setup.service Service Type oneshot

# The Core listens on loopback only, so it must not need IPv6.
if ! section_lines "$units/luna-core.service" Service | grep -qxF 'Environment=LUNA_LISTEN=127.0.0.1:7766'; then
  fail "luna-core.service must listen on 127.0.0.1 only"
fi

if command -v systemd-analyze >/dev/null && [[ -e /.dockerenv || -e /run/.containerenv ]]; then
  # Fake binaries only satisfy ExecStart; safe because this is a container.
  for bin in luna-core luna-policy-broker luna-owner-setup; do
    if [[ ! -e /usr/bin/$bin ]]; then
      printf '#!/bin/bash\nexit 0\n' > "/usr/bin/$bin"
      chmod 755 "/usr/bin/$bin"
    fi
  done
  # One call so the Requires=/After= units resolve against each other.
  if ! systemd-analyze verify "$units/luna-owner-setup.service" "$units/luna-policy-broker.service" "$units/luna-core.service"; then
    fail "systemd-analyze verify"
  fi
  for unit in luna-core.service luna-policy-broker.service luna-owner-setup.service; do
    systemd-analyze security --offline=yes --no-pager "$units/$unit" || true
  done
else
  echo "SKIP: systemd-analyze verify needs a disposable Linux container"
fi

if (( failures > 0 )); then
  echo "systemd-units-test: $failures failure(s)" >&2
  exit 1
fi
echo "systemd-units-test: ok"
