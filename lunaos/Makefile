SHELL := /bin/bash

.PHONY: test check build dev test-reference check-reference smoke-native native-smoke

test:
	cargo test --workspace --locked --manifest-path rust/Cargo.toml

check:
	cargo fmt --all --manifest-path rust/Cargo.toml -- --check
	cargo clippy --workspace --all-targets --locked --manifest-path rust/Cargo.toml -- -D warnings
	cargo check --workspace --locked --manifest-path rust/Cargo.toml

build:
	cargo build --release --locked --manifest-path rust/Cargo.toml -p luna-core -p luna-policy-broker

dev:
	cargo run --locked --manifest-path rust/Cargo.toml -p luna-core

smoke-native: build
	./scripts/smoke-native-core.sh

native-smoke: smoke-native

# Frozen Go/Python behavioral reference only. These tools are not required by Core.
test-reference:
	go test -buildvcs=false -count=1 ./services/policy
	python3 -m unittest discover -s tests -v

check-reference:
	go vet -buildvcs=false ./services/policy
	python3 -m compileall -q luna scripts tests
	bash -n scripts/linux-baseline.sh
