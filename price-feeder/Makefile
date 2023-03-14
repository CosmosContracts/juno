BRANCH    := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DIR ?= $(CURDIR)/build
COMMIT    := $(shell git log -1 --format='%H')
SDK_VERSION     := $(shell go list -m github.com/cosmos/cosmos-sdk | sed 's:.* ::')

all: test-unit install

.PHONY: all

###############################################################################
##                                  Version                                  ##
###############################################################################

ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

###############################################################################
##                              Build / Install                              ##
###############################################################################

ldflags = -X github.com/CosmosContracts/juno/price-feeder/cmd.Version=$(VERSION) \
		  -X github.com/CosmosContracts/juno/price-feeder/cmd.Commit=$(COMMIT) \
		  -X github.com/CosmosContracts/juno/price-feeder/cmd.SDKVersion=$(SDK_VERSION)

ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif

build_tags += $(BUILD_TAGS)

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

build: go.sum
	@echo "--> Building..."
	go build -mod=readonly -o $(BUILD_DIR)/ $(BUILD_FLAGS) ./...

install: go.sum
	@echo "--> Installing..."
	go install -mod=readonly $(BUILD_FLAGS) ./...

tidy:	
	go mod tidy

.PHONY: build install

###############################################################################
##                              Tests & Linting                              ##
###############################################################################

test-unit:
	@echo "--> Running tests"
	@go test -mod=readonly -race ./... -v

.PHONY: test-unit

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout=8m

.PHONY: lint
