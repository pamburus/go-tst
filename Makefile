# Common makefile helpers
include build/make/common.mk

# Common configuration
SHELL := $(SHELL) -o pipefail

# Set default goal
.DEFAULT_GOAL := all

# Populate complete module list, including build tools
ifndef all-modules
all-modules := $(shell go list -m -f '{{.Dir}}')
all-modules := $(all-modules:$(PWD)/%=%)
all-modules := $(all-modules:$(PWD)=.)
endif

# Auxiliary modules, not to be tested
aux-modules := 

# Populate module list to test
ifndef modules
modules := $(filter-out $(aux-modules),$(all-modules))
endif

# Tools
go-test := go test
ifeq ($(verbose),yes)
	go-test += -v
endif
 
## Run all tests
.PHONY: all
all: ci

# ---

## Run continuous integration tests
.PHONY: ci
ci: lint test

.PHONY: ci/%
ci/%: lint/% test/%

# ---

## Run linters
.PHONY: lint
lint: $(modules:%=lint/%)

.PHONY: lint/%
lint/%:
	golangci-lint run $*/...

# ---

## Run tests
.PHONY: test
test: $(modules:%=test/%)

.PHONY: test/%
test/%:
	$(go-test) ./$*/...
test/.:
	$(go-test) ./...

# ---

## Tidy up
.PHONY: tidy
tidy: $(all-modules:%=tidy/%)

.PHONY: tidy/%
tidy/%:
	cd $* && go mod tidy

# ---

## Clean up
.PHONY: clean
clean:
	find . -type f -name go.work.sum -delete
