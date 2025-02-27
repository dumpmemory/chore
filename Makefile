APP_NAME := chore
GOBIN    := .bin

GOLANGCI_LINT_VERSION := v1.50.1
GOFUMPT_VERSION       := v0.4.0

VERSION      := $(shell git describe --exact-match HEAD 2>/dev/null || git describe --tags --always)
STATIC_FLAGS := -trimpath -mod=readonly -ldflags="-s -w -X 'main.version=$(VERSION)'"
GOTOOL       := env "GOBIN=$(abspath $(GOBIN))" "PATH=$(abspath $(GOBIN)):$(PATH)"
GO_FILES     := $(shell find . -name "*.go" -type f | grep -vE '_test\.go$$')

# -----------------------------------------------------------------------------

.PHONY: all
all: $(APP_NAME)

$(APP_NAME): $(GO_FILES) go.sum
	@go build -tags timetzdata -o "$(APP_NAME)"

vendor: go.mod go.sum
	@$(MOD_ON) go mod vendor

$(GOBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)/golangci-lint:
	@env GOBIN=$(abspath $(dir $@)) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(GOBIN)/golangci-lint: $(GOBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)/golangci-lint
	@ln -sf $(abspath $<) $@

$(GOBIN)/gofumpt-$(GOFUMPT_VERSION)/gofumpt:
	@env GOBIN=$(abspath $(dir $@)) go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)

$(GOBIN)/gofumpt: $(GOBIN)/gofumpt-$(GOFUMPT_VERSION)/gofumpt
	@ln -sf $(abspath $<) $@

# -----------------------------------------------------------------------------

.PHONY: static
static:
	@env go build \
		$(STATIC_FLAGS) \
		-tags netgo \
		-tags timetzdata \
		-a \
		-o "$(APP_NAME)"

.PHONY: test
test:
	@go test -v -parallel 4 ./...

.PHONY: full-test
full-test:
	@go test -v -parallel 4 -race -cover -coverprofile coverage.out ./...

.PHONY: lint
lint: $(GOBIN)/golangci-lint
	@$(GOTOOL) golangci-lint run ./...

.PHONY: fmt
fmt: $(GOBIN)/gofumpt
	@$(GOTOOL) gofumpt -extra -w .

.PHONY: clean
clean:
	@git clean -xfd

.PHONY: update
update:
	@go get -u && go mod tidy -go=1.19
