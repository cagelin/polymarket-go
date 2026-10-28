GOLANGCI_LINT ?= golangci-lint

.PHONY: check lint test race fmt

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
