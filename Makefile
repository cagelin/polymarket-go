GOLANGCI_LINT ?= golangci-lint
PYTHON ?= python3

.PHONY: check lint test race fmt generate

check: lint test race

lint:
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) fmt --diff
	$(GOLANGCI_LINT) run ./...

test:
	CGO_ENABLED=0 go test ./...

race:
	go test -race ./...

fmt:
	$(GOLANGCI_LINT) fmt

generate:
	GOLANGCI_LINT=$(GOLANGCI_LINT) $(PYTHON) scripts/regenerate.py
