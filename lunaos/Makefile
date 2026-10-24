PYTHON ?= python3

.PHONY: env test check dev

env:
	$(PYTHON) -m venv --without-pip .venv

test:
	go test -buildvcs=false -count=1 ./services/policy
	$(PYTHON) -m unittest discover -s tests -v

check:
	go vet -buildvcs=false ./services/policy
	$(PYTHON) -m compileall -q luna scripts tests
	bash -n scripts/linux-baseline.sh

dev:
	$(PYTHON) scripts/dev.py
