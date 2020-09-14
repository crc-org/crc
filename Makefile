SHELL := /bin/bash

BUNDLE_VERSION = 4.5.9
# OC_VERSION and BUNDLE_VERSION are going to same for release artifacts but
# different for nightly and CI bits where bundle version would be any random
# string or dd-mm-yyyy format.
OC_VERSION ?= ${BUNDLE_VERSION}
BUNDLE_EXTENSION = crcbundle
CRC_VERSION = 1.16.0
COMMIT_SHA=$(shell git rev-parse --short HEAD)

# Go and compilation related variables
BUILD_DIR ?= out
SOURCE_DIRS = cmd pkg test
RELEASE_DIR ?= release

# Docs build related variables
DOCS_BUILD_DIR ?= docs/build
DOCS_BUILD_CONTAINER ?= quay.io/crcont/docs-builder:latest
DOCS_BUILD_TARGET ?= /docs/source/getting_started/master.adoc

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
HOST_BUILD_DIR=$(BUILD_DIR)/$(GOOS)-$(GOARCH)
GOPATH ?= $(shell go env GOPATH)
ORG := github.com/code-ready
REPOPATH ?= $(ORG)/crc

SOURCES := $(shell git ls-files  *.go ":^vendor")

RELEASE_INFO := release-info.json

MOCK_BUNDLE ?= false

# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# Params:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

# Linker flags
VERSION_VARIABLES := -X $(REPOPATH)/pkg/crc/version.crcVersion=$(CRC_VERSION) \
    -X $(REPOPATH)/pkg/crc/version.bundleVersion=$(BUNDLE_VERSION) \
    -X $(REPOPATH)/pkg/crc/version.ocVersion=$(OC_VERSION) \
	-X $(REPOPATH)/pkg/crc/version.commitSha=$(COMMIT_SHA)

# https://golang.org/cmd/link/
LDFLAGS := $(VERSION_VARIABLES) -extldflags='-static' -s -w

# Add default target
.PHONY: default
default: install

# Create and update the vendor directory
.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

.PHONY: vendorcheck
vendorcheck:
	./verify-vendor.sh

# Start of the actual build targets

.PHONY: install
install: $(SOURCES)
	go install -ldflags="$(LDFLAGS)" ./cmd/crc

$(BUILD_DIR)/macos-amd64/crc: $(SOURCES)
	GOARCH=amd64 GOOS=darwin go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/macos-amd64/crc ./cmd/crc

$(BUILD_DIR)/linux-amd64/crc: $(SOURCES)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/crc ./cmd/crc

$(BUILD_DIR)/windows-amd64/crc.exe: $(SOURCES)
	GOARCH=amd64 GOOS=windows go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/crc.exe ./cmd/crc

$(HOST_BUILD_DIR)/crc-embedder: $(SOURCES)
	go build --tags="build" -ldflags="$(LDFLAGS)" -o $(HOST_BUILD_DIR)/crc-embedder ./cmd/crc-embedder

.PHONY: cross ## Cross compiles all binaries
cross: $(BUILD_DIR)/macos-amd64/crc $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: test
test:
	go test --tags build -v -ldflags="$(VERSION_VARIABLES)" ./pkg/... ./cmd/...

.PHONY: build_docs
build_docs:
	podman run -v $(CURDIR)/docs:/docs:Z --rm $(DOCS_BUILD_CONTAINER) build_docs -b html5 -D /$(DOCS_BUILD_DIR) -o index.html $(DOCS_BUILD_TARGET)

.PHONY: build_docs_pdf
build_docs_pdf:
	podman run -v $(CURDIR)/docs:/docs:Z --rm $(DOCS_BUILD_CONTAINER) build_docs_pdf -D /$(DOCS_BUILD_DIR) -o doc.pdf $(DOCS_BUILD_TARGET)

.PHONY: docs_serve
docs_serve: build_docs
	podman run -it -v $(CURDIR)/docs:/docs:Z --rm -p 8088:8088/tcp $(DOCS_BUILD_CONTAINER) docs_serve 

.PHONY: docs_check_links
docs_check_links:
	podman run -it -v $(CURDIR)/docs:/docs:Z --rm $(DOCS_BUILD_CONTAINER) docs_check_links

.PHONY: clean_docs
clean_docs:
	rm -rf $(CURDIR)/docs/build

.PHONY: clean ## Remove all build artifacts
clean: clean_docs
	rm -rf $(BUILD_DIR)
	rm -f $(GOPATH)/bin/crc
	rm -rf $(RELEASE_DIR)

.PHONY: build_integration
build_integration: $(SOURCES)
	GOOS=linux   go test ./test/integration/ -c -o $(BUILD_DIR)/linux-amd64/integration.test
	GOOS=windows go test ./test/integration/ -c -o $(BUILD_DIR)/windows-amd64/integration.test.exe
	GOOS=darwin  go test ./test/integration/ -c -o $(BUILD_DIR)/macos-amd64/integration.test

.PHONY: integration ## Run integration tests
integration:
GODOG_OPTS = --godog.tags=$(GOOS)
ifndef PULL_SECRET_FILE
	PULL_SECRET_FILE = --pull-secret-file=$(HOME)/Downloads/crc-pull-secret
endif
ifndef BUNDLE_LOCATION
	BUNDLE_LOCATION = --bundle-location=$(HOME)/Downloads/crc_libvirt_$(BUNDLE_VERSION).$(BUNDLE_EXTENSION)
endif
ifndef CRC_BINARY
	CRC_BINARY = --crc-binary=$(GOPATH)/bin
endif
integration:
	@go test --timeout=120m $(REPOPATH)/test/integration -v $(PULL_SECRET_FILE) $(BUNDLE_LOCATION) $(CRC_BINARY) --bundle-version=$(BUNDLE_VERSION) $(GODOG_OPTS)

.PHONY: fmt
fmt:
	@gofmt -l -w $(SOURCE_DIRS)

$(GOPATH)/bin/golangci-lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0

# Run golangci-lint against code
.PHONY: lint cross-lint
lint: $(GOPATH)/bin/golangci-lint
	$(GOPATH)/bin/golangci-lint run

cross-lint: $(GOPATH)/bin/golangci-lint
	GOOS=darwin $(GOPATH)/bin/golangci-lint run
	GOOS=linux $(GOPATH)/bin/golangci-lint run
	GOOS=windows $(GOPATH)/bin/golangci-lint run

.PHONY: gen_release_info
gen_release_info:
	@cat release-info.json.sample | sed s/@CRC_VERSION@/\"$(CRC_VERSION)\"/ > $(RELEASE_INFO)
	@sed -i s/@GIT_COMMIT_SHA@/\"$(COMMIT_SHA)\"/ $(RELEASE_INFO)
	@sed -i s/@OPENSHIFT_VERSION@/\"$(BUNDLE_VERSION)\"/ $(RELEASE_INFO)

.PHONY: release
release: cross-lint embed_bundle build_docs_pdf gen_release_info
	mkdir $(RELEASE_DIR)
	
	@mkdir -p $(BUILD_DIR)/crc-macos-$(CRC_VERSION)-amd64
	@cp LICENSE $(DOCS_BUILD_DIR)/doc.pdf $(BUILD_DIR)/macos-amd64/crc $(BUILD_DIR)/crc-macos-$(CRC_VERSION)-amd64
	tar cJSf $(RELEASE_DIR)/crc-macos-amd64.tar.xz -C $(BUILD_DIR) crc-macos-$(CRC_VERSION)-amd64 --owner=0 --group=0

	@mkdir -p $(BUILD_DIR)/crc-linux-$(CRC_VERSION)-amd64
	@cp LICENSE $(DOCS_BUILD_DIR)/doc.pdf $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/crc-linux-$(CRC_VERSION)-amd64
	tar cJSf $(RELEASE_DIR)/crc-linux-amd64.tar.xz -C $(BUILD_DIR) crc-linux-$(CRC_VERSION)-amd64 --owner=0 --group=0
	
	@mkdir -p $(BUILD_DIR)/crc-windows-$(CRC_VERSION)-amd64
	@cp LICENSE $(DOCS_BUILD_DIR)/doc.pdf $(BUILD_DIR)/windows-amd64/crc.exe $(BUILD_DIR)/crc-windows-$(CRC_VERSION)-amd64
	cd $(BUILD_DIR) && zip -r $(CURDIR)/$(RELEASE_DIR)/crc-windows-amd64.zip crc-windows-$(CRC_VERSION)-amd64

	@mv $(RELEASE_INFO) $(RELEASE_DIR)/$(RELEASE_INFO)
	
	pushd $(RELEASE_DIR) && sha256sum * > sha256sum.txt && popd

HYPERKIT_BUNDLENAME = $(BUNDLE_DIR)/crc_hyperkit_$(BUNDLE_VERSION).$(BUNDLE_EXTENSION)
HYPERV_BUNDLENAME = $(BUNDLE_DIR)/crc_hyperv_$(BUNDLE_VERSION).$(BUNDLE_EXTENSION)
LIBVIRT_BUNDLENAME = $(BUNDLE_DIR)/crc_libvirt_$(BUNDLE_VERSION).$(BUNDLE_EXTENSION)

.PHONY: embed_bundle check_bundledir
check_bundledir:
ifeq ($(MOCK_BUNDLE),true)
	touch $(HYPERKIT_BUNDLENAME) $(HYPERV_BUNDLENAME) $(LIBVIRT_BUNDLENAME)
endif
	@$(call check_defined, BUNDLE_DIR, "Embedding bundle requires BUNDLE_DIR set to a directory containing CRC bundles for all hypervisors")

embed_bundle: clean cross $(HOST_BUILD_DIR)/crc-embedder check_bundledir $(HYPERKIT_BUNDLENAME) $(HYPERV_BUNDLENAME) $(LIBVIRT_BUNDLENAME)
	$(HOST_BUILD_DIR)/crc-embedder embed --log-level debug --goos=darwin --bundle-dir=$(BUNDLE_DIR) $(BUILD_DIR)/macos-amd64/crc
	$(HOST_BUILD_DIR)/crc-embedder embed --log-level debug --goos=linux --bundle-dir=$(BUNDLE_DIR) $(BUILD_DIR)/linux-amd64/crc
	$(HOST_BUILD_DIR)/crc-embedder embed --log-level debug --goos=windows --bundle-dir=$(BUNDLE_DIR) $(BUILD_DIR)/windows-amd64/crc.exe
