#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BINDIR ?= $(GOPATH)/bin
APP = ./app

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BFT_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::') # grab everything after the space in "github.com/cometbft/cometbft v0.34.7"
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.0.0-rc8
BUILDDIR ?= $(CURDIR)/build
E2E_UPGRADE_VERSION := "v14"
export GO111MODULE = on

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

ifeq (cleveldb,$(findstring cleveldb,$(JUNO_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

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
			-X github.com/cometbft/cometbft/version.TMCoreSemVer=$(BFT_VERSION)

ifeq (cleveldb,$(findstring cleveldb,$(JUNO_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ifeq (,$(findstring nostrip,$(JUNO_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(JUNO_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif
 
#$(info $$BUILD_FLAGS is [$(BUILD_FLAGS)])
include contrib/devtools/Makefile

all: install
	@echo "--> project root: go mod tidy"	
	@go mod tidy			
	@echo "--> project root: linting --fix"	
	@GOGC=1 golangci-lint run --fix --timeout=8m

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/junod

build:
	go build $(BUILD_FLAGS) -o bin/junod ./cmd/junod

test-node:
	CHAIN_ID="local-1" HOME_DIR="~/.juno1" TIMEOUT_COMMIT="500ms" CLEAN=true sh scripts/test_node.sh

###############################################################################
###                                Testing                                  ###
###############################################################################

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(APP) -ExitOnFail 50 10 TestFullAppSimulation

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_UNIT)

###############################################################################
###                             e2e interchain test                         ###
###############################################################################

# Executes basic chain tests via interchaintest
ictest-basic: rm-testcache
	cd interchaintest && go test -race -v -run TestBasicJunoStart .

ictest-ibchooks: rm-testcache
	cd interchaintest && go test -race -v -run TestJunoIBCHooks .

ictest-tokenfactory: rm-testcache
	cd interchaintest && go test -race -v -run TestJunoTokenFactory .

ictest-feeshare: rm-testcache
	cd interchaintest && go test -race -v -run TestJunoFeeShare .

ictest-pfm: rm-testcache
	cd interchaintest && go test -race -v -run TestPacketForwardMiddlewareRouter .

# Executes a basic chain upgrade test via interchaintest
ictest-upgrade: rm-testcache
	cd interchaintest && go test -race -v -run TestBasicJunoUpgrade .

# Executes a basic chain upgrade locally via interchaintest after compiling a local image as juno:local
ictest-upgrade-local: local-image ictest-upgrade

# Executes IBC tests via interchaintest
ictest-ibc: rm-testcache
	cd interchaintest && go test -race -v -run TestJunoGaiaIBCTransfer .

# Unity contract CI
ictest-unity-deploy: rm-testcache
	cd interchaintest && go test -race -v -run TestJunoUnityContractDeploy .

ictest-unity-gov: rm-testcache
	cd interchaintest && go test -race -v -run TestJunoUnityContractGovSubmit .

rm-testcache:
	go clean -testcache

.PHONY: test-mutation ictest-basic ictest-upgrade ictest-ibc ictest-unity-deploy ictest-unity-gov

###############################################################################
###                                  heighliner                             ###
###############################################################################

get-heighliner:
	git clone https://github.com/strangelove-ventures/heighliner.git
	cd heighliner && go install

local-image:
ifeq (,$(shell which heighliner))
	echo 'heighliner' binary not found. Consider running `make get-heighliner`
else
	heighliner build -c juno --local -f ./chains.yaml
endif

.PHONY: get-heighliner local-image

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.13.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@$(protoImage) sh ./scripts/protoc-swagger-gen.sh
	$(MAKE) update-swagger-docs

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

.PHONY: proto-all proto-gen proto-gen-any proto-swagger-gen proto-format proto-lint proto-check-breaking proto-update-deps docs
