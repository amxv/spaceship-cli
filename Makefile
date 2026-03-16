GO ?= go
NODE ?= npm
DIST_DIR ?= dist
BIN_NAME ?= spaceship
BIN_PATH := $(DIST_DIR)/$(BIN_NAME)

.PHONY: help fmt typecheck test vet staticcheck lint-go lint-node lint check build install-local smoke clean release-tag

help:
	@echo "spaceship-domains-cli command runner"
	@echo ""
	@echo "Quality:"
	@echo "  make fmt          - gofmt source files"
	@echo "  make typecheck    - compile all Go packages"
	@echo "  make test         - run go tests"
	@echo "  make vet          - run go vet"
	@echo "  make staticcheck  - run staticcheck (if installed)"
	@echo "  make lint-go      - run golangci-lint (if installed)"
	@echo "  make lint-node    - run npm lint script"
	@echo "  make check        - run full local quality pass"
	@echo ""
	@echo "Build/Run:"
	@echo "  make build        - build CLI binary to dist/spaceship"
	@echo "  make install-local- install CLI to ~/.local/bin/spaceship"
	@echo "  make smoke        - run built binary help output"
	@echo ""
	@echo "Release:"
	@echo "  make release-tag TAG=vX.Y.Z - push release tag to trigger CI release workflow"

fmt:
	$(GO) fmt ./...

typecheck:
	$(GO) build ./...

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

staticcheck:
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not found. Install: brew install staticcheck OR go install honnef.co/go/tools/cmd/staticcheck@latest"; \
		exit 1; \
	fi

lint-go:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install: brew install golangci-lint"; \
		exit 1; \
	fi

lint-node:
	$(NODE) run lint

lint: lint-go lint-node

check: fmt typecheck test vet staticcheck lint-go lint-node

build:
	mkdir -p $(DIST_DIR)
	$(GO) build -o $(BIN_PATH) ./cmd/spaceship

install-local: build
	mkdir -p $$HOME/.local/bin
	install -m 755 $(BIN_PATH) $$HOME/.local/bin/spaceship

smoke: build
	./$(BIN_PATH) --help

clean:
	rm -rf $(DIST_DIR)

release-tag:
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make release-tag TAG=vX.Y.Z"; \
		exit 1; \
	fi
	git tag $(TAG)
	git push origin $(TAG)
