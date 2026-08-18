#!/usr/bin/make -f

# set variables
HOST_GOOS := $(shell go env GOOS 2>/dev/null)
HTTPS_GIT := $(shell git config --get remote.origin.url)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags --always 2>/dev/null)
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
ifneq ($(strip $(BUILD_TAGS)),)
  build_tags += $(BUILD_TAGS)
endif
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
  ifeq ($(HOST_GOOS),linux)
    ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
  endif
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -trimpath -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################

verify:
	@echo "🔎 Verifying Dependencies ..."
	@go mod verify > /dev/null 2>&1
	@echo "✅ Verified dependencies successfully!"

go-cache: verify
	@echo "📥 Downloading and caching dependencies..."
	@go mod download
	@echo "✅ Downloaded and cached dependencies successfully!"

install: go-cache
	@echo "🔄 Installing Juno..."
	@go install $(BUILD_FLAGS) -mod=readonly ./cmd/junod
	@echo "✅ Installed Juno successfully! Run it using 'junod'!"
	@echo ""
	@echo "====== Install Summary ======"
	@echo "Juno: $(VERSION)"
	@echo "Cosmos SDK: $(COSMOS_SDK_VERSION)"
	@echo "Comet: $(CMT_VERSION)"
	@echo "============================="

build: go-cache
	@echo "🔄 Building Juno..."
	@if [ "$(OS)" = "Windows_NT" ]; then \
		GOOS=windows GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o bin/junod.exe ./cmd/junod; \
	else \
		go build -mod=readonly $(BUILD_FLAGS) -o bin/junod ./cmd/junod; \
	fi
	@echo "✅ Built Juno successfully! Run it using './bin/junod'!"
	@echo ""
	@echo "======= Build Summary ======="
	@echo "Juno: $(VERSION)"
	@echo "Cosmos SDK: $(COSMOS_SDK_VERSION)"
	@echo "Comet: $(CMT_VERSION)"
	@echo "============================="

init:
	sh scripts/init.sh

.PHONY: verify tidy go-cache install build init

###############################################################################
###                                 Tooling                                 ###
###############################################################################

lint:
	@echo "🔄 Linting code..."
	@go tool golangci-lint run --config ./.golangci.yml
	@echo "✅ Linted code successfully!"

format:
	@echo "🔄 Formatting code..."
	@go tool gofumpt -l -w .
	@echo "✅ Formatted code successfully!"

.PHONY: format lint

###############################################################################
###                                 E2E Tests                               ###
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
	cd interchaintest/tests/upgrade && go test -race -v -run TestUpgradeTestSuite .

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

ictest-dao-dao: rm-testcache
	cd interchaintest/tests/dao-dao && go test -race -v -run TestDaoDaoTestSuite .

rm-testcache:
	go clean -testcache

.PHONY: ictest-basic ictest-cw ictest-node ictest-feemarket ictest-fees ictest-upgrade ictest-ibc ictest-ibc-hooks ictest-pfm ictest-tokenfactory ictest-drip ictest-burn ictest-fixes ictest-dao-dao rm-testcache

###############################################################################
###                                Docker                                   ###
###############################################################################

IMAGE ?= ghcr.io/cosmoscontracts/juno
PLATFORMS ?= linux/amd64,linux/arm64
BUILDER ?= multiarch

UNAME_ARCH := $(shell uname -m)
ifeq ($(UNAME_ARCH),x86_64)
  LOCAL_PLATFORM ?= linux/amd64
else ifeq ($(UNAME_ARCH),arm64)
  LOCAL_PLATFORM ?= linux/arm64
else ifeq ($(UNAME_ARCH),aarch64)
  LOCAL_PLATFORM ?= linux/arm64
else
  LOCAL_PLATFORM ?= linux/amd64
endif

setup-builder:
	@$(DOCKER) buildx inspect $(BUILDER) >/dev/null 2>&1 || \
		$(DOCKER) buildx create --name $(BUILDER) --driver docker-container --use
	@$(DOCKER) buildx use $(BUILDER)
	@$(DOCKER) buildx inspect --bootstrap

local-image: setup-builder
	@echo "🔄 Building Docker Image..."
	$(DOCKER) buildx build \
		--load \
		--platform=$(LOCAL_PLATFORM) \
		-t $(IMAGE):local \
		-f Dockerfile \
		.
	@echo "✅ Built Docker Image successfully!"

proto-image: setup-builder
	@echo "🔄 Building Protobuilder Image..."
	$(DOCKER) buildx build \
		--load \
		--platform=$(LOCAL_PLATFORM) \
		-t $(PROTO_IMAGE_NAME) \
		-f proto/Dockerfile .
	@echo "✅ Built Proto Image successfully!"

.PHONY: setup-builder local-image proto-image

###############################################################################
###                                Protobuf                                 ###
###############################################################################

PROTO_IMAGE_NAME := juno-protobuilder:latest
PROTO_IMAGE := $(DOCKER) run --rm -v "$(CURDIR)":/workspace --workdir /workspace $(PROTO_IMAGE_NAME)

proto-all: proto-check proto-gen
proto-gen: proto-gogo proto-pulsar proto-openapi
proto-check: proto-format proto-lint

proto-gogo:
	@echo "🛠️ Generating Gogo types from Protobuffers"
	@$(PROTO_IMAGE) sh ./scripts/buf/buf-gogo.sh
	@echo "✅ Generated Gogo types successfully!"

proto-pulsar:
	@echo "🛠️ Generating Pulsar types from Protobuffers"
	@$(PROTO_IMAGE) sh ./scripts/buf/buf-pulsar.sh
	@echo "✅ Generated Pulsar types successfully!"

proto-openapi:
	@echo "🛠️ Generating OpenAPI Spec from Protobuffers"
	@$(PROTO_IMAGE) sh ./scripts/buf/buf-openapi.sh
	@echo "✅ Generated OpenAPI Spec successfully!"

proto-format:
	@echo "🖊️ Formatting Protobuffers"
	@$(PROTO_IMAGE) go tool buf format ./proto --error-format=json
	@echo "✅ Formatted Protobuffers successfully!"

proto-lint:
	@echo "🔎 Linting Protobuffers"
	@$(PROTO_IMAGE) go tool buf lint --error-format=json
	@echo "✅ Linted Protobuffers successfully!"

proto-breaking:
	@echo "🔎 Checking breaking Protobuffers changes against branch main"
	@$(PROTO_IMAGE) buf breaking ./proto --against $(HTTPS_GIT).git#branch=main
	@echo "✅ Protobuffers are non-breaking, checked successfully!"

.PHONY: proto-all proto-gen proto-check proto-format proto-lint proto-breaking proto-gogo proto-pulsar proto-openapi
