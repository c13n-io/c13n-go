.PHONY: clean distclean fmt tidy vendor proto proto-doc mock dev test testlib btcd build-itest lib-utest lib-itest testbackend lint all

TARGET := c13n

MODULE_NAME = github.com/c13n-io/c13n-go

SERVICE_DIR = rpc/services

PROTO_IMPORT_PATHS := --proto_path=. --proto_path=vendor

GO = go

############
# Packages #
############

BTCD_PKG := github.com/btcsuite/btcd
BTCD_PKG_VERSION := v0.21.0-beta.0.20210513141527-ee5896bad5be
LND_PKG := github.com/lightningnetwork/lnd
LND_PKG_VERSION := v0.13.1-beta

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
RPC_DOCS_FILE = rpc.md

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

LDFLAGS := -ldflags "-X $(MODULE_NAME)/app.commit=$(COMMIT) \
	-X $(MODULE_NAME)/app.commitHash=$(COMMIT_HASH)"

$(TARGET):
	$(GO) build -i -v -o $(TARGET) $(LDFLAGS) $(MODULE_NAME)/cli

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
	(cd && GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@v1.4.3)
	(cd && GO111MODULE=on go get github.com/mwitkow/go-proto-validators/...@v0.3.0)
	(cd && GO111MODULE=on go get github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.3.2)
	(cd && GO111MODULE=on go get golang.org/x/tools/cmd/goimports)
	(cd && GO111MODULE=on go get -u github.com/mgechev/revive@v1.0.2)
	(cd && GO111MODULE=on go get github.com/vektra/mockery/...@v1.0.0)
	(cd && GO111MODULE=on go get honnef.co/go/tools/cmd/staticcheck@v0.2.0)

# Generate protobuf source code
# http://github.com/golang/protobuf
proto: vendor
	protoc $(PROTO_IMPORT_PATHS) --go_out=plugins=grpc,paths=source_relative:. --govalidators_out=paths=source_relative:. $(SERVICE_DIR)/*.proto

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

testlib: btcd build-itest lib-utest lib-itest

build-itest:
	@echo "Installing lnd and btcd."
	$(GO) get $(LND_PKG)@$(LND_PKG_VERSION) $(BTCD_PKG)@$(BTCD_PKG_VERSION)
	@echo "Building itest lnd and lncli."
	$(GO) build -mod mod -tags="$(ITEST_TAGS)" -o ./lnchat/tests/lnd-itest $(LND_PKG)/cmd/lnd
	$(GO) build -mod mod -tags="$(ITEST_TAGS)" -o ./lnchat/tests/lncli-itest $(LND_PKG)/cmd/lncli

lib-itest:
	@echo "Running integration tests with btcd backend."
	$(RM) ./lnchat/tests/*.log
	$(GO) test -v -coverprofile=lnchat/tests/coverage.out -coverpkg=$(ITEST_COVERAGE_PKGS) ./lnchat/tests -tags="$(ITEST_TAGS)" -logoutput

lib-utest:
	@echo "Running lnchat unit tests."
	$(GO) test -count=1 $(LNCHAT_PKG)

# Linting
lint:
	@echo "Running revive"
	revive -config=revive.toml -formatter=stylish $(C13N_PACKAGES)
	@echo "Running staticcheck"
	staticcheck ./...

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
