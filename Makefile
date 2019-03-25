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

CRC_VERSION = 0.0.1-alpha.1
COMMIT_SHA=$(shell git rev-parse --short HEAD)

# Go and compilation related variables
BUILD_DIR ?= out

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ORG := github.com/code-ready
REPOPATH ?= $(ORG)/crc
ifeq ($(GOOS),windows)
	IS_EXE := .exe
endif

PACKAGES := go list ./... | grep -v /out

# Linker flags
VERSION_VARIABLES := -X $(REPOPATH)/pkg/version.crcVersion=$(CRC_VERSION) \
	-X $(REPOPATH)/pkg/version.commitSha=$(COMMIT_SHA)

LDFLAGS := $(VERSION_VARIABLES)

# Start of the actual build targets

.PHONY: $(CURDIR)/bin/crc$(IS_EXE)
$(CURDIR)/bin/crc$(IS_EXE):
	go install -ldflags="$(VERSION_VARIABLES)" ./cmd/crc

$(BUILD_DIR)/$(GOOS)-$(GOARCH):
	mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)

$(BUILD_DIR)/darwin-amd64/crc: $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	GOARCH=amd64 GOOS=darwin go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/darwin-amd64/crc ./cmd/crc

$(BUILD_DIR)/linux-amd64/crc:  $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/crc ./cmd/crc

$(BUILD_DIR)/windows-amd64/crc.exe:  $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	GOARCH=amd64 GOOS=windows go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/crc.exe ./cmd/crc

.PHONY: cross ## Cross compiles all binaries
cross: $(BUILD_DIR)/darwin-amd64/crc $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: test
test:
	@go test -v -ldflags="$(VERSION_VARIABLES)" $(shell $(PACKAGES))

.PHONY: clean ## Remove all build artifacts
clean:
	rm -rf $(BUILD_DIR)

