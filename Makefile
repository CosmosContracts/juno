#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

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
TM_VERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::') # grab everything after the space in "github.com/tendermint/tendermint v0.34.7"
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build

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
			-X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TM_VERSION)

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


all: install

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/junod

build:
	go build $(BUILD_FLAGS) -o bin/junod ./cmd/junod

###############################################################################
###                                 Localnet                                ###
###############################################################################
localnet-keys:
	. tests/localjuno/scripts/add_keys.sh

localnet-init: localnet-clean localnet-build

localnet-build:  
	@chmod -R +x tests/localjuno/
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f tests/localjuno/docker-compose.yml build

localnet-start:  
	@STATE="" docker-compose -f tests/localjuno/docker-compose.yml up

localnet-start-with-state:	
	@STATE=-s docker-compose -f tests/localjuno/docker-compose.yml up

localnet-startd:
	@STATE="" docker-compose -f tests/localjuno/docker-compose.yml up -d

localnet-startd-with-state:
	@STATE=-s docker-compose -f tests/localjuno/docker-compose.yml up -d

localnet-stop:
	@STATE="" docker-compose -f tests/localjuno/docker-compose.yml down

localnet-clean:
	@rm -rfI $(HOME)/.juno/

localnet-state-export-init: localnet-state-export-clean localnet-state-export-build 

localnet-state-export-build:
	@chmod -R +x tests/localjuno/
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f tests/localjuno/state_export/docker-compose.yml build

localnet-state-export-start:
	@docker-compose -f tests/localjuno/state_export/docker-compose.yml up

localnet-state-export-startd:
	@docker-compose -f tests/localjuno/state_export/docker-compose.yml up -d

localnet-state-export-stop:
	@docker-compose -f tests/localjuno/docker-compose.yml down

localnet-state-export-clean: localnet-clean

###############################################################################
###                                  Proto                                  ###
###############################################################################

protoVer=v0.7
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=juno-proto-gen-$(protoVer)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi
