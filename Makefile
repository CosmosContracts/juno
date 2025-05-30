#!/usr/bin/make -f

# set variables
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif
LEDGER_ENABLED ?= true
COSMOS_SDK_VERSION := $(shell go list -m github.com/cosmos/cosmos-sdk | sed 's:.* ::')
CMT_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
DOCKER := $(shell which docker)

# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
	ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
   endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=juno \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=junod \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(CMT_VERSION)

ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################

verify:
	@echo "🔎 - Verifying Dependencies ..."
	@go mod verify > /dev/null 2>&1
	@go mod tidy
	@echo "✅ - Verified dependencies successfully!"
	@echo ""

go-cache: verify
	@echo "📥 - Downloading and caching dependencies..."
	@go mod download
	@echo "✅ - Downloaded and cached dependencies successfully!"
	@echo ""

install: go-cache
	@echo "🔄 - Installing Juno..."
	@go install $(BUILD_FLAGS) -mod=readonly ./cmd/junod
	@echo "✅ - Installed Juno successfully! Run it using 'junod'!"
	@echo ""
	@echo "====== Install Summary ======"
	@echo "Juno: $(VERSION)"
	@echo "Cosmos SDK: $(COSMOS_SDK_VERSION)"
	@echo "Comet: $(CMT_VERSION)"
	@echo "============================="

build: go-cache
	@echo "🔄 - Building Juno..."
	@if [ "$(OS)" = "Windows_NT" ]; then \
		GOOS=windows GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o bin/junod.exe ./cmd/junod; \
	else \
		go build -mod=readonly $(BUILD_FLAGS) -o bin/junod ./cmd/junod; \
	fi
	@echo "✅ - Built Juno successfully! Run it using './bin/junod'!"
	@echo ""
	@echo "====== Install Summary ======"
	@echo "Juno: $(VERSION)"
	@echo "Cosmos SDK: $(COSMOS_SDK_VERSION)"
	@echo "Comet: $(CMT_VERSION)"
	@echo "============================="

test-node:
	CHAIN_ID="local-1" HOME_DIR="~/.juno" TIMEOUT_COMMIT="500ms" CLEAN=true sh scripts/test_node.sh

.PHONY: verify go-cache install build test-node

###############################################################################
###                                 Tooling                                 ###
###############################################################################

gofumpt=mvdan.cc/gofumpt
gofumpt_version=v0.8.0

golangci_lint=github.com/golangci/golangci-lint/v2/cmd/golangci-lint
golangci_lint_version=v2.1.6

install-format:
	@echo "🔄 - Installing gofumpt $(gofumpt_version)..."
	@go install $(gofumpt)@$(gofumpt_version)
	@echo "✅ - Installed gofumpt successfully!"
	@echo ""

install-lint:
	@echo "🔄 - Installing golangci-lint $(golangci_lint_version)..."
	@go install $(golangci_lint)@$(golangci_lint_version)
	@echo "✅ - Installed golangci-lint successfully!"
	@echo ""

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		INSTALLED=$$(golangci-lint version | head -n1 | awk '{print $$4}'); \
		echo "Detected golangci-lint $$INSTALLED, required $(golangci_lint_version)"; \
		if [ "$$(printf '%s\n' "$(golangci_lint_version)" "$$INSTALLED" | sort -V | head -n1)" != "$(golangci_lint_version)" ]; then \
	   	echo "Updating golangci-lint..."; \
	   	$(MAKE) install-lint; \
		fi; \
	else \
		echo "golangci-lint not found; installing..."; \
		$(MAKE) install-lint; \
	fi
	@echo "🔄 - Linting code..."
	@golangci-lint run
	@echo "✅ - Linted code successfully!"

format:
	@if command -v gofumpt >/dev/null 2>&1; then \
		INSTALLED=$$(go version -m $$(command -v gofumpt) | awk '$$1=="mod" {print $$3; exit}'); \
		echo "Detected gofumpt $$INSTALLED, required $(gofumpt_version)"; \
		if [ "$$(printf '%s\n' "$(gofumpt_version)" "$$INSTALLED" | sort -V | head -n1)" != "$(gofumpt_version)" ]; then \
	   	echo "Updating gofumpt..."; \
	   	$(MAKE) install-format; \
		fi; \
	else \
		echo "gofumpt not found; installing..."; \
		$(MAKE) install-format; \
	fi
	@echo "🔄 - Formatting code..."
	@gofumpt -l -w .
	@echo "✅ - Formatted code successfully!"

.PHONY: install-format format install-lint lint

###############################################################################
###                             e2e interchain test                         ###
###############################################################################

ictest-basic: rm-testcache
	cd interchaintest/tests/basic && go test -race -v -run TestBasicTestSuite .

ictest-cw: rm-testcache
	cd interchaintest/tests/cosmwasm && go test -race -v -run TestCosmWasmTestSuite .

ictest-node: rm-testcache
	cd interchaintest/tests/node && go test -race -v -run TestNodeTestSuite .

ictest-feemarket: rm-testcache
	cd interchaintest/tests/feemarket && go test -race -v -run TestFeemarketTestSuite .

ictest-fees: rm-testcache
	cd interchaintest/tests/fees && go test -race -v -run TestFeesTestSuite .

ictest-upgrade: rm-testcache
	cd interchaintest/tests/upgrade && go test -race -v -run BasicUpgradeTestSuite .

ictest-ibc: rm-testcache
	cd interchaintest/tests/ibc && go test -race -v -run TestIbcTestSuite .

ictest-ibc-hooks: rm-testcache
	cd interchaintest/tests/ibc-hooks && go test -race -v -run TestIbcHooksTestSuite .

ictest-pfm: rm-testcache
	cd interchaintest/tests/pfm && go test -race -v -run TestPfmTestSuite .

ictest-tokenfactory: rm-testcache
	cd interchaintest/tests/tokenfactory && go test -race -v -run TestTokenfactoryTestSuite .

ictest-drip: rm-testcache
	cd interchaintest/tests/drip && go test -race -v -run TestDripTestSuite .

ictest-burn: rm-testcache
	cd interchaintest/tests/burn && go test -race -v -run TestBurnTestSuite .

ictest-fixes: rm-testcache
	cd interchaintest/tests/fixes && go test -race -v -run TestFixTestSuite .

rm-testcache:
	go clean -testcache

.PHONY: ictest-basic ictest-cw ictest-node ictest-fees ictest-upgrade ictest-ibc ictest-tokenfactory ictest-drip ictest-burn ictest-drip ictest-burn ictest-fixes rm-testcache

###############################################################################
###                                  heighliner                             ###
###############################################################################

heighliner=github.com/strangelove-ventures/heighliner
heighliner_version=v1.7.2

install-heighliner:
	@if ! command -v heighliner > /dev/null; then \
   	echo "🔄 - Installing heighliner $(heighliner_version)..."; \
      go install $(heighliner)@$(heighliner_version); \
		echo "✅ - Installed heighliner successfully!"; \
		echo ""; \
   fi

local-image: install-heighliner
	@echo "🔄 - Building Docker Image..."
	heighliner build --chain juno --local -f ./chains.yaml
	@echo "✅ - Built Docker Image successfully!"

.PHONY: install-heighliner local-image

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.17.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace -v /var/run/docker.sock:/var/run/docker.sock --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-check-breaking proto-gogo proto-pulsar proto-openapi

proto-gogo:
	@echo "🛠️ - Generating Gogo types from Protobuffers"
	@$(protoImage) sh ./scripts/buf/buf-gogo.sh
	@echo "✅ - Generated Gogo types successfully!"

proto-pulsar:
	@echo "🛠️ - Generating Pulsar types from Protobuffers"
	@$(protoImage) sh ./scripts/buf/buf-pulsar.sh
	@echo "✅ - Generated Pulsar types successfully!"

proto-openapi:
	@echo "🛠️ - Generating OpenAPI Spec from Protobuffers"
	@sh ./scripts/buf/buf-openapi.sh
	@echo "✅ - Generated OpenAPI Spec successfully!"

proto-format:
	@echo "🖊️ - Formatting Protobuffers"
	@$(protoImage) buf format ./proto --error-format=json
	@echo "✅ - Formatted Protobuffers successfully!"

proto-lint:
	@echo "🔎 - Linting Protobuffers"
	@$(protoImage) buf lint --error-format=json
	@echo "✅ - Linted Protobuffers successfully!"

proto-check-breaking:
	@echo "🔎 - Checking breaking Protobuffers changes against branch main"
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main
	@echo "✅ - Protobuffers are non-breaking, checked successfully!"

.PHONY: proto-all proto-format proto-lint proto-check-breaking proto-gogo proto-pulsar proto-openapi
