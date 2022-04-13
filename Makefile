.PHONY: clean distclean fmt tidy vendor proto proto-doc mock dev test testlib btcd build-itest lib-utest lib-itest testbackend lint all

TARGET := c13n

MODULE_NAME = github.com/c13n-io/c13n-go
SHELL = /bin/bash

include make/release_flags.mk

SERVICE_DIR = rpc/services

PROTO_IMPORT_PATHS := --proto_path=. --proto_path=vendor

GO = go

GOBUILD := go build -v
GOINSTALL := go install

############
# Packages #
############

BTCD_PKG := github.com/btcsuite/btcd
LND_PKG := github.com/lightningnetwork/lnd

# Backend Packages
BACKEND_PACKAGES := $(shell go list ./... | grep -vE "lnchat")

# Library packages
LNCHAT_PKG := github.com/c13n-io/c13n-go/lnchat
LNCONNECT_PKG := github.com/c13n-io/c13n-go/lnchat/lnconnect

C13N_PACKAGES := $(BACKEND_PACKAGES) $(LNCHAT_PKG) $(LNCONNECT_PKG)

#############
# Artifacts #
#############

# Backend certificates
CERT_DIR = cert

# Documentation
RPC_DOCS_FILE = index.md

############
# Testing #
###########

# Integration testing
## Integration test flags
ITEST_TAGS := dev rpctest chainrpc walletrpc signrpc invoicesrpc autopilotrpc routerrpc watchtowerrpc wtclientrpc btcd

## Packages to include in coverage reports, comma separated
ITEST_COVERAGE_PKGS := $(LNCHAT_PKG),$(LNCONNECT_PKG)

# Test artifacts
## Test output logs regex
TEST_OUTPUT_LOGS := lnchat/tests/*.log

## Test output logs regex
TEST_HARNESS_LOGS := lnchat/tests/.backendlogs lnchat/tests/.minerlogs

## Coverage output file
COVERAGE_OUTPUT := lnchat/tests/coverage.out

## LND test executable
LND_TEST_EXEC := lnchat/tests/lnd-itest

## LNCLI test executable
LNCLI_TEST_EXEC := lnchat/tests/lncli-itest

############
# Targets #
###########

all: $(TARGET)
dev: proto mock tidy

COMMIT := $(shell git describe --abbrev=40 --dirty)
COMMIT_HASH := $(shell git rev-parse HEAD)

LDFLAGSBASE := -X $(MODULE_NAME)/app.commit=$(COMMIT) \
	-X $(MODULE_NAME)/app.commitHash=$(COMMIT_HASH)

LDFLAGS := -ldflags="$(LDFLAGSBASE)"

RELEASE_LDFLAGS := -s -w -buildid= $(LDFLAGSBASE)

$(TARGET):
	$(GOBUILD) -o $(TARGET) $(LDFLAGS) $(MODULE_NAME)/cli

release:
	./scripts/release.sh build-release "$(BUILD_SYSTEM)" "$(RELEASE_LDFLAGS)" $(MODULE_NAME)/cli

certgen:
	openssl req -nodes -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -config $(CERT_DIR)/cert.conf -extensions v3_exts -days 365 -keyout $(CERT_DIR)/c13n.key -out $(CERT_DIR)/c13n.pem

# Formatting
fmt:
	goimports -e -w -local $(MODULE_NAME) .

tidy: fmt
	$(GO) mod tidy

# Building
# https://stepan.wtf/importing-protobuf-with-go-modules/
vendor:
	go mod vendor

dev-deps:
	$(GOINSTALL) google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0
	$(GOINSTALL) google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	$(GOINSTALL) github.com/mwitkow/go-proto-validators/...@v0.3.2
	$(GOINSTALL) github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.3.2
	$(GOINSTALL) github.com/vektra/mockery/v2@v2.10.4
	$(GOINSTALL) golang.org/x/tools/cmd/goimports@latest
	$(GOINSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.1

# Generate protobuf source code
# http://github.com/golang/protobuf
proto: vendor
	protoc $(PROTO_IMPORT_PATHS) \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--govalidators_out=. --govalidators_opt=paths=source_relative \
		$(SERVICE_DIR)/*.proto

# Generating
mock:
	$(GO) generate ./...

proto-doc: vendor
	protoc $(PROTO_IMPORT_PATHS) --doc_out=./docs/c13n-api-docs/docs --doc_opt=./docs/md.template,$(RPC_DOCS_FILE) $(SERVICE_DIR)/*.proto

# Testing
test: testlib testbackend

testbackend:
	@echo "Executing backend tests"
	$(GO) test -count=1 $(BACKEND_PACKAGES)

testlib: build-itest lib-utest lib-itest

build-itest:
	@echo "Building itest lnd and lncli."
	$(GOBUILD) -mod=mod -tags="$(ITEST_TAGS)" -o ./lnchat/tests/lnd-itest $(LND_PKG)/cmd/lnd
	$(GOBUILD) -mod=mod -tags="$(ITEST_TAGS)" -o ./lnchat/tests/lncli-itest $(LND_PKG)/cmd/lncli

lib-itest:
	@echo "Running integration tests with btcd backend."
	$(RM) ./lnchat/tests/*.log
	$(GO) test -v -coverprofile=lnchat/tests/coverage.out -coverpkg=$(ITEST_COVERAGE_PKGS) ./lnchat/tests -tags="$(ITEST_TAGS)" -logoutput

lib-utest:
	@echo "Running lnchat unit tests."
	$(GO) test -count=1 $(LNCHAT_PKG)

# Linting
lint:
	@echo "Running linters"
	golangci-lint run ./...

# Cleaning
clean:
	$(RM) $(TARGET)
	# Remove lnd and lncli executables
	$(RM) $(LND_TEST_EXEC) $(LNCLI_TEST_EXEC)
	# Remove logs from btcd, lnd-itest, lncli-itest
	$(RM) $(TEST_OUTPUT_LOGS)
	# Remove harness logs
	$(RM) -r $(TEST_HARNESS_LOGS)
	# Remove go test coverage output
	$(RM) $(COVERAGE_OUTPUT)

distclean: clean
	$(RM) -r vendor/
	$(RM) $(CERT_DIR)/c13n.pem $(CERT_DIR)/c13n.key
	$(RM) docs/$(RPC_DOCS_FILE)
	$(RM) -r c13n-build/
