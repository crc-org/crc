# Copyright (C) 2019 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BUNDLE_VERSION = 4.1.14
BUNDLE_EXTENSION = crcbundle
CRC_VERSION = 1.0.0-dev
COMMIT_SHA=$(shell git rev-parse --short HEAD)

# Go and compilation related variables
BUILD_DIR ?= out
SOURCE_DIRS = cmd pkg test
RELEASE_DIR ?= release

# Docs build related variables
DOCS_BUILD_DIR ?= docs/build
DOCS_BUILD_CONTAINER ?= registry.gitlab.com/gbraad/asciidoctor-centos:latest
DOCS_PDF_CONTAINER ?= quay.io/crcont/asciidoctor-pdf:latest
DOCS_SERVE_CONTAINER ?= quay.io/crcont/docs-server:latest
DOCS_BUILD_TARGET ?= /docs/source/getting-started/master.adoc

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)
ORG := github.com/code-ready
REPOPATH ?= $(ORG)/crc

PACKAGES := go list ./... | grep -v /out
SOURCES := $(shell git ls-files  *.go ":^vendor")

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
	-X $(REPOPATH)/pkg/crc/version.commitSha=$(COMMIT_SHA)

BUNDLE_EMBEDDED := -X $(REPOPATH)/pkg/crc/constants.bundleEmbedded=true

# https://golang.org/cmd/link/
LDFLAGS := $(VERSION_VARIABLES) -extldflags='-static' -s -w

# Add default target
.PHONY: default
default: install

# Create and update the vendor directory
.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor

# Get binappend
$(GOPATH)/bin/binappend-cli:
	GO111MODULE=off go get -u github.com/yourfin/binappend-cli

# Start of the actual build targets

.PHONY: install
install: $(SOURCES)
	go install -ldflags="$(VERSION_VARIABLES)" ./cmd/crc

$(BUILD_DIR)/macos-amd64/crc: $(SOURCES)
	GOARCH=amd64 GOOS=darwin go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/macos-amd64/crc ./cmd/crc

$(BUILD_DIR)/linux-amd64/crc: $(SOURCES)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/crc ./cmd/crc

$(BUILD_DIR)/windows-amd64/crc.exe: $(SOURCES)
	GOARCH=amd64 GOOS=windows go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/crc.exe ./cmd/crc

.PHONY: cross ## Cross compiles all binaries
cross: $(BUILD_DIR)/macos-amd64/crc $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: test
test:
	go test -v -ldflags="$(VERSION_VARIABLES)" $(shell $(PACKAGES))

.PHONY: build_docs
build_docs:
	podman run -v $(CURDIR)/docs:/docs:Z --rm $(DOCS_BUILD_CONTAINER) -b html5 -D /$(DOCS_BUILD_DIR) -o index.html $(DOCS_BUILD_TARGET)

.PHONY: build_docs_pdf
build_docs_pdf:
	podman run -v $(CURDIR)/docs:/docs:Z --rm $(DOCS_PDF_CONTAINER) -D /$(DOCS_BUILD_DIR) -o doc.pdf $(DOCS_BUILD_TARGET)

.PHONY: docs_serve
docs_serve: build_docs
	podman run -v $(CURDIR)/docs:/docs:Z --rm -p 8088:8088/tcp $(DOCS_SERVE_CONTAINER)

.PHONY: clean_docs
clean_docs:
	rm -rf $(CURDIR)/docs/build

.PHONY: clean ## Remove all build artifacts
clean: clean_docs
	rm -rf $(BUILD_DIR)
	rm -f $(GOPATH)/bin/crc
	rm -rf $(RELEASE_DIR)
       

.PHONY: integration ## Run integration tests
integration: GODOG_OPTS = --godog.tags=$(GOOS)
integration:
	@$(call check_defined, BUNDLE_LOCATION, "'make integration' requires BUNDLE_LOCATION to contain the full path to a bundle file")
	@$(call check_defined, PULL_SECRET_FILE, "'make integration' requires PULL_SECRET_FILE to point to a file with the pull secret to use")
	@go test --timeout=60m $(REPOPATH)/test/integration -v --tags=integration $(GODOG_OPTS) $(BUNDLE_LOCATION) $(PULL_SECRET_FILE)

.PHONY: fmt
fmt:
	@gofmt -l -w $(SOURCE_DIRS)

.PHONY: fmtcheck
fmtcheck: ## Checks for style violation using gofmt
	@gofmt -l $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

$(GOPATH)/bin/golangci-lint:
	GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.1

# Run golangci-lint against code
.PHONY: lint cross-lint
lint: $(GOPATH)/bin/golangci-lint
	$(GOPATH)/bin/golangci-lint run

cross-lint: $(GOPATH)/bin/golangci-lint
	GOOS=darwin $(GOPATH)/bin/golangci-lint run
	GOOS=linux $(GOPATH)/bin/golangci-lint run
	GOOS=windows $(GOPATH)/bin/golangci-lint run

$(GOPATH)/bin/gosec:
	GO111MODULE=on go get github.com/securego/gosec/cmd/gosec@2.0.0

# Run gosec against code
gosec: $(GOPATH)/bin/gosec
	$(GOPATH)/bin/gosec -tests -severity medium  ./...

.PHONY: release
release: fmtcheck embed_bundle build_docs_pdf
	mkdir $(RELEASE_DIR)
	
	@mkdir -p $(BUILD_DIR)/crc-macos-$(CRC_VERSION)-amd64
	@cp LICENSE $(DOCS_BUILD_DIR)/doc.pdf $(BUILD_DIR)/macos-amd64/crc $(BUILD_DIR)/crc-macos-$(CRC_VERSION)-amd64
	tar cJSf $(RELEASE_DIR)/crc-macos-amd64.tar.xz -C $(BUILD_DIR) crc-macos-$(CRC_VERSION)-amd64
	sha256sum $(RELEASE_DIR)/crc-macos-amd64.tar.xz > $(RELEASE_DIR)/sha256sum.txt

	@mkdir -p $(BUILD_DIR)/crc-linux-$(CRC_VERSION)-amd64
	@cp LICENSE $(DOCS_BUILD_DIR)/doc.pdf $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/crc-linux-$(CRC_VERSION)-amd64
	tar cJSf $(RELEASE_DIR)/crc-linux-amd64.tar.xz -C $(BUILD_DIR) crc-linux-$(CRC_VERSION)-amd64
	sha256sum $(RELEASE_DIR)/crc-linux-amd64.tar.xz >> $(RELEASE_DIR)/sha256sum.txt
	
	@mkdir -p $(BUILD_DIR)/crc-windows-$(CRC_VERSION)-amd64
	@cp LICENSE $(DOCS_BUILD_DIR)/doc.pdf $(BUILD_DIR)/windows-amd64/crc.exe $(BUILD_DIR)/crc-windows-$(CRC_VERSION)-amd64
	cd $(BUILD_DIR) && zip -r $(CURDIR)/$(RELEASE_DIR)/crc-windows-amd64.zip crc-windows-$(CRC_VERSION)-amd64
	sha256sum $(RELEASE_DIR)/crc-windows-amd64.zip >> $(RELEASE_DIR)/sha256sum.txt

HYPERKIT_BUNDLENAME = $(BUNDLE_DIR)/crc_hyperkit_$(BUNDLE_VERSION).$(BUNDLE_EXTENSION)
HYPERV_BUNDLENAME = $(BUNDLE_DIR)/crc_hyperv_$(BUNDLE_VERSION).$(BUNDLE_EXTENSION)
LIBVIRT_BUNDLENAME = $(BUNDLE_DIR)/crc_libvirt_$(BUNDLE_VERSION).$(BUNDLE_EXTENSION)

.PHONY: embed_bundle check_bundledir
check_bundledir:
	@$(call check_defined, BUNDLE_DIR, "Embedding bundle requires BUNDLE_DIR set to a directory containing CRC bundles for all hypervisors")

embed_bundle: LDFLAGS += $(BUNDLE_EMBEDDED)
embed_bundle: clean cross $(GOPATH)/bin/binappend-cli check_bundledir $(HYPERKIT_BUNDLENAME) $(HYPERV_BUNDLENAME) $(LIBVIRT_BUNDLENAME)
	$(GOPATH)/bin/binappend-cli write $(BUILD_DIR)/linux-amd64/crc $(LIBVIRT_BUNDLENAME)
	$(GOPATH)/bin/binappend-cli write $(BUILD_DIR)/macos-amd64/crc $(HYPERKIT_BUNDLENAME)
	$(GOPATH)/bin/binappend-cli write $(BUILD_DIR)/windows-amd64/crc.exe $(HYPERV_BUNDLENAME)
