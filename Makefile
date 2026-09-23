BIN := hush
PKG := ./...

.PHONY: check fmt vet lint test build hook

check: fmt vet test build

fmt:
	gofmt -l .
	@test -z "$$(gofmt -l .)" || (echo "gofmt: archivos sin formatear" && exit 1)

vet:
	go vet $(PKG)

lint:
	golangci-lint run $(PKG)

test:
	go test $(PKG)

build:
	go build -o $(BIN) ./cmd/hush

hook:
	git config core.hooksPath .githooks
	@echo "hook instalado: .githooks/pre-commit"
