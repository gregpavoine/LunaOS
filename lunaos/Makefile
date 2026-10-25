SHELL := /bin/bash

.PHONY: test test-process-linux check check-qml build build-linux-amd64 build-linux-amd64-v3 build-local test-reference check-reference test-rust-checkpoint check-rust-checkpoint smoke-rust-checkpoint

test:
	go test -buildvcs=false -count=1 ./services/policy ./services/core/... ./services/intimacy/... ./services/identity/...

# Real Linux child-process smoke: isolated broker, Core, SQLite recovery and API.
test-process-linux:
	@if [[ "$$(uname -s)" != Linux ]]; then echo 'SKIP: process integration requires Linux Unix-socket peer credentials' >&2; exit 2; fi
	go test -buildvcs=false -tags integration -count=1 -run '^TestGoCoreAndPolicyBrokerProcessesRecoverQueuedGoal$$' ./services/core

check:
	test -z "$$(gofmt -l $$(find services -name '*.go' -type f))"
	go vet -buildvcs=false ./services/policy ./services/core/... ./services/intimacy/... ./services/identity/...
	bash -n scripts/linux-baseline.sh scripts/build-archiso.sh scripts/dev-native.sh packaging/arch/PKGBUILD
	python3 -m json.tool contracts/openapi/local.json >/dev/null

# Run from an Arch/Qt environment that provides Quickshell's qmllint tooling.
check-qml:
	@if command -v qmllint >/dev/null 2>&1; then \
		qmllint shell/shell.qml; \
	elif [[ -x /usr/lib/qt6/bin/qmllint ]]; then \
		/usr/lib/qt6/bin/qmllint shell/shell.qml; \
	else \
		echo 'Install Quickshell/Qt declarative tooling to run qmllint.' >&2; exit 2; \
	fi

build-local:
	mkdir -p build/local
	CGO_ENABLED=1 go build -buildvcs=false -trimpath -o build/local/luna-policy-broker ./services/policy
	CGO_ENABLED=1 go build -buildvcs=false -trimpath -o build/local/luna-core ./services/core/cmd/luna-core
	CGO_ENABLED=1 go build -buildvcs=false -trimpath -o build/local/luna-vault ./services/intimacy/cmd/luna-vault
	CGO_ENABLED=1 go build -buildvcs=false -trimpath -o build/local/luna-owner-setup ./services/identity/cmd/luna-owner-setup

build: build-local

# Generic Arch x86_64 compatibility baseline for the broker. The Core uses the
# Cgo SQLite driver and must be built in an Arch environment with its native toolchain.
build-linux-amd64:
	mkdir -p build/target
	GOOS=linux GOARCH=amd64 GOAMD64=v1 CGO_ENABLED=0 go build -buildvcs=false -trimpath -o build/target/luna-policy-broker-x86_64 ./services/policy

# Ryzen 9 5950X target profile. Install only after CPU feature detection confirms v3.
build-linux-amd64-v3:
	mkdir -p build/target
	GOOS=linux GOARCH=amd64 GOAMD64=v3 CGO_ENABLED=0 go build -buildvcs=false -trimpath -o build/target/luna-policy-broker-x86_64-v3 ./services/policy

# Frozen Go/Python behavioral reference only. These tools are not required by Core.
test-reference:
	go test -buildvcs=false -count=1 ./services/policy
	@if [[ "$$(uname -s)" == Linux ]]; then python3 -m unittest discover -s tests -v; else echo 'SKIP: frozen Python process tests require Linux SO_PEERCRED'; fi

check-reference:
	go vet -buildvcs=false ./services/policy
	python3 -m compileall -q luna scripts tests
	bash -n scripts/linux-baseline.sh

test-rust-checkpoint:
	cargo test --workspace --locked --manifest-path rust/Cargo.toml

check-rust-checkpoint:
	cargo fmt --all --manifest-path rust/Cargo.toml -- --check
	cargo clippy --workspace --all-targets --locked --manifest-path rust/Cargo.toml -- -D warnings
	cargo check --workspace --locked --manifest-path rust/Cargo.toml

smoke-rust-checkpoint:
	cargo build --release --locked --manifest-path rust/Cargo.toml -p luna-core -p luna-policy-broker -p luna-identity
	./scripts/smoke-native-core.sh
