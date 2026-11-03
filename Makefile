GOLANGCI_LINT ?= golangci-lint
PYTHON ?= python3

.PHONY: verify-generated check lint test race fmt generate validate-release coverage

check: verify-generated lint test race coverage

verify-generated:
	$(PYTHON) scripts/generate_exports.py --check

lint:
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) fmt --diff
	$(GOLANGCI_LINT) run ./...

test:
	CGO_ENABLED=0 go test ./...
	$(PYTHON) -m unittest discover -s tests/acceptance -p 'test_*.py'

race:
	go test -race ./...

fmt:
	$(GOLANGCI_LINT) fmt

generate:
	GOLANGCI_LINT=$(GOLANGCI_LINT) $(PYTHON) scripts/regenerate.py
	$(MAKE) lint

validate-release: check
	GOLANGCI_LINT=$(GOLANGCI_LINT) $(PYTHON) scripts/validate_release.py all

coverage:
	sh scripts/check_coverage.sh
