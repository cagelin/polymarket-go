GOLANGCI_LINT ?= golangci-lint
PYTHON ?= python3
SHADOW_TOOL := $(CURDIR)/.local/tools/check-shadow.exe

.PHONY: check lint shadow test race fmt generate

check: lint test race

lint: shadow
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) fmt --diff
	$(GOLANGCI_LINT) run ./...
	cd tools/shadow && $(GOLANGCI_LINT) fmt --diff --config ../../.golangci.yml
	cd tools/shadow && $(GOLANGCI_LINT) run --config ../../.golangci.yml ./...

shadow:
	go -C tools/shadow build -o "$(SHADOW_TOOL)" .
	go vet -vettool="$(SHADOW_TOOL)" ./...
	go -C tools/shadow vet -vettool="$(SHADOW_TOOL)" ./...

test:
	CGO_ENABLED=0 go test ./...
	CGO_ENABLED=0 go -C tools/shadow test ./...

race:
	go test -race ./...
	go -C tools/shadow test -race ./...

fmt:
	$(GOLANGCI_LINT) fmt
	cd tools/shadow && $(GOLANGCI_LINT) fmt --config ../../.golangci.yml

generate:
	GOLANGCI_LINT=$(GOLANGCI_LINT) $(PYTHON) scripts/regenerate.py
	$(MAKE) shadow
