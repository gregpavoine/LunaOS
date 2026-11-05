SHELL := /bin/bash

.PHONY: test check test-process-linux build build-local

test:
	go test -buildvcs=false -count=1 ./services/...

check:
	test -z "$$(gofmt -l services)"
	go vet -buildvcs=false ./services/...
	bash -n scripts/dev-native.sh
	for f in $$(find contracts -name '*.json' -type f); do python3 -m json.tool "$$f" >/dev/null || exit 1; done

# Real Linux child-process smoke: isolated broker, Core, SQLite recovery and API.
test-process-linux:
	@if [[ "$$(uname -s)" != Linux ]]; then echo 'SKIP: process integration requires Linux Unix-socket peer credentials' >&2; exit 2; fi
	go test -buildvcs=false -tags integration -count=1 -run '^TestGoCoreAndPolicyBrokerProcessesRecoverQueuedGoal$$' ./services/core

build: build-local

build-local:
	mkdir -p build/local
	CGO_ENABLED=1 go build -buildvcs=false -trimpath -o build/local/luna-policy-broker ./services/policy
	CGO_ENABLED=1 go build -buildvcs=false -trimpath -o build/local/luna-core ./services/core/cmd/luna-core
	CGO_ENABLED=1 go build -buildvcs=false -trimpath -o build/local/luna-vault ./services/intimacy/cmd/luna-vault
	CGO_ENABLED=1 go build -buildvcs=false -trimpath -o build/local/luna-owner-setup ./services/identity/cmd/luna-owner-setup
