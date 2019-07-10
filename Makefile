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

BUNDLE_VERSION = 4.1.3
CRC_VERSION = 0.88.0-alpha
COMMIT_SHA=$(shell git rev-parse --short HEAD)

# Go and compilation related variables
BUILD_DIR ?= out
SOURCE_DIRS = cmd pkg test

# Docs build related variables
DOCS_BUILD_DIR ?= /docs/build
DOCS_BUILD_CONTAINER ?= registry.gitlab.com/gbraad/asciidoctor-centos:latest
DOCS_BUILD_TARGET ?= /docs/source/getting-started/master.adoc

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ORG := github.com/code-ready
REPOPATH ?= $(ORG)/crc
ifeq ($(GOOS),windows)
	IS_EXE := .exe
endif

PACKAGES := go list ./... | grep -v /out

# Bundle embedding related variables
BUNDLE_ASSET_DIR ?= $(CURDIR)/bundle_bindata
BUNDLE_ASSET_FILE ?= $(BUNDLE_ASSET_DIR)/bundle_data.go

# Linker flags
VERSION_VARIABLES := -X $(REPOPATH)/pkg/crc/version.crcVersion=$(CRC_VERSION) \
    -X $(REPOPATH)/pkg/crc/version.bundleVersion=$(BUNDLE_VERSION) \
	-X $(REPOPATH)/pkg/crc/version.commitSha=$(COMMIT_SHA)

ifeq ($(EMBED_BUNDLE), true) # Embeds crc bundle in the binary
	ASSET_TAG:=-tags embed_bundle
	VERSION_VARIABLES := $(VERSION_VARIABLES) -X $(REPOPATH)/pkg/crc/constants.bundleEmbedded=true
endif

# https://golang.org/cmd/link/
LDFLAGS := $(VERSION_VARIABLES) -extldflags='-static' -s -w

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

# Add default target
.PHONY: default
default: $(CURDIR)/bin/crc$(IS_EXE)

# Create and update the vendor directory
.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor

# Get go-bindata
bindata:
	GO111MODULE=off go get -u github.com/go-bindata/go-bindata/...

# Generate go files for assets (bundle)
assets: bindata
	@:$(call check_defined, BUNDLE_LOCATION, "BUNDLE_LOCATION needs to be set and pointing to a CRC bundle")
	@mkdir -p $(BUNDLE_ASSET_DIR)
	go-bindata -nocompress -tags "embed_bundle" -o $(BUNDLE_ASSET_FILE) -prefix $(BUNDLE_ASSET_DIR) -pkg bundle_bindata $(BUNDLE_LOCATION)

# Start of the actual build targets

.PHONY: $(CURDIR)/bin/crc$(IS_EXE)
$(CURDIR)/bin/crc$(IS_EXE):
	go install -ldflags="$(VERSION_VARIABLES)" $(ASSET_TAG) ./cmd/crc

$(BUILD_DIR)/$(GOOS)-$(GOARCH):
	mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)

$(BUILD_DIR)/darwin-amd64/crc: $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	GOARCH=amd64 GOOS=darwin go build -ldflags="$(LDFLAGS)" $(ASSET_TAG) -o $(BUILD_DIR)/darwin-amd64/crc ./cmd/crc

$(BUILD_DIR)/linux-amd64/crc:  $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" $(ASSET_TAG) -o $(BUILD_DIR)/linux-amd64/crc ./cmd/crc

$(BUILD_DIR)/windows-amd64/crc.exe:  $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	GOARCH=amd64 GOOS=windows go build -ldflags="$(LDFLAGS)" $(ASSET_TAG) -o $(BUILD_DIR)/windows-amd64/crc.exe ./cmd/crc

.PHONY: cross ## Cross compiles all binaries
cross: $(BUILD_DIR)/darwin-amd64/crc $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: package ## build crc binaries with embedded bundle
package: assets $(BUILD_DIR)/darwin-amd64/crc $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: test
test:
	go test -v -ldflags="$(VERSION_VARIABLES)" $(shell $(PACKAGES))

.PHONY: build_docs
build_docs:
	podman run -v $(CURDIR)/docs:/docs:Z --rm $(DOCS_BUILD_CONTAINER) -b html5 -D $(DOCS_BUILD_DIR) $(DOCS_BUILD_TARGET)

.PHONY: clean_docs
clean_docs:
	rm -rf $(CURDIR)/docs/build

.PHONY: clean ## Remove all build artifacts
clean: clean_docs
	rm -rf $(BUNDLE_ASSET_FILE)
	rm -rf $(BUILD_DIR)
	rm -f $(GOPATH)/bin/crc

.PHONY: integration ## Run integration tests
integration: GODOG_OPTS = --godog.tags=$(GOOS)
integration:
	@:$(call check_defined, BUNDLE_LOCATION, "Running integration tests requires BUNDLE_LOCATION set and pointing to a CRC bundle")
	go test --timeout=60m $(REPOPATH)/test/integration -v --tags=integration $(GODOG_OPTS) $(BUNDLE_LOCATION)

.PHONY: fmt
fmt:
	@gofmt -l -w $(SOURCE_DIRS)

.PHONY: fmtcheck
fmtcheck: ## Checks for style violation using gofmt
	@gofmt -l $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

