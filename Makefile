SHELL := /bin/bash

OPENSHIFT_VERSION ?= 4.10.4
PODMAN_VERSION ?= 3.4.4
BUNDLE_EXTENSION = crcbundle
CRC_VERSION = 2.0.1
COMMIT_SHA=$(shell git rev-parse --short HEAD)
MACOS_INSTALL_PATH = /Applications/CodeReady Containers.app/Contents/Resources
CONTAINER_RUNTIME ?= podman
GOLANGCI_LINT_VERSION = v1.44.2

ifdef OKD_VERSION
    OPENSHIFT_VERSION = $(OKD_VERSION)
    CRC_VERSION := $(CRC_VERSION)-OKD
endif

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
PACKAGE_DIR := packaging/$(GOOS)

SOURCES := $(shell git ls-files  *.go ":^vendor")

RELEASE_INFO := release-info.json

CUSTOM_EMBED ?= false
EMBED_DOWNLOAD_DIR ?= tmp-embed

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
	-X $(REPOPATH)/pkg/crc/version.bundleVersion=$(OPENSHIFT_VERSION) \
	-X $(REPOPATH)/pkg/crc/version.podmanVersion=$(PODMAN_VERSION) \
	-X $(REPOPATH)/pkg/crc/version.commitSha=$(COMMIT_SHA)
RELEASE_VERSION_VARIABLES := -X $(REPOPATH)/pkg/crc/segment.WriteKey=cvpHsNcmGCJqVzf6YxrSnVlwFSAZaYtp

ifdef OKD_VERSION
	VERSION_VARIABLES := $(VERSION_VARIABLES) -X $(REPOPATH)/pkg/crc/version.okdBuild=true
endif

# https://golang.org/cmd/link/
LDFLAGS := $(VERSION_VARIABLES) ${GO_EXTRA_LDFLAGS}

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

.PHONY: check
check: cross build_e2e $(HOST_BUILD_DIR)/crc-embedder test cross-lint vendorcheck

# Start of the actual build targets

.PHONY: install
install: $(SOURCES)
	go install -ldflags="$(LDFLAGS)" $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(BUILD_DIR)/macos-amd64/crc: $(SOURCES)
	GOARCH=amd64 GOOS=darwin go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/macos-amd64/crc $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(BUILD_DIR)/linux-amd64/crc: $(SOURCES)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/crc $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(BUILD_DIR)/windows-amd64/crc.exe: $(SOURCES)
	GOARCH=amd64 GOOS=windows go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/crc.exe $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(HOST_BUILD_DIR)/crc-embedder: $(SOURCES)
	go build --tags="build" -ldflags="$(LDFLAGS)" -o $(HOST_BUILD_DIR)/crc-embedder $(GO_EXTRA_BUILDFLAGS) ./cmd/crc-embedder

.PHONY: cross ## Cross compiles all binaries
cross: $(BUILD_DIR)/macos-amd64/crc $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: containerized ## Cross compile from container
containerized: clean
	${CONTAINER_RUNTIME} build -t crc-build -f images/build .
	${CONTAINER_RUNTIME} run --name crc-cross crc-build make cross
	${CONTAINER_RUNTIME} cp crc-cross:/opt/app-root/src/out ./
	${CONTAINER_RUNTIME} rm crc-cross
	${CONTAINER_RUNTIME} rmi crc-build

.PHONY: test
test:
	go test -race --tags build -v -ldflags="$(VERSION_VARIABLES)" ./pkg/... ./cmd/...

.PHONY: spec test-rpmbuild

GENERATED_RPM_FILES=packaging/rpm/crc.spec images/rpmbuild/Containerfile
spec: $(GENERATED_RPM_FILES)
	
test-rpmbuild: spec
	${CONTAINER_RUNTIME} build -f images/rpmbuild/Containerfile .

.PHONY: build_docs
build_docs:
	${CONTAINER_RUNTIME} run -v $(CURDIR)/docs:/docs:Z --rm $(DOCS_BUILD_CONTAINER) build_docs -b html5 -D /$(DOCS_BUILD_DIR) -o index.html $(DOCS_BUILD_TARGET)

.PHONY: docs_serve
docs_serve: build_docs
	${CONTAINER_RUNTIME} run -it -v $(CURDIR)/docs:/docs:Z --rm -p 8088:8088/tcp $(DOCS_BUILD_CONTAINER) docs_serve 

.PHONY: docs_check_links
docs_check_links:
	${CONTAINER_RUNTIME} run -it -v $(CURDIR)/docs:/docs:Z --rm $(DOCS_BUILD_CONTAINER) docs_check_links

.PHONY: clean_docs clean_macos_package
clean_docs:
	rm -rf $(CURDIR)/docs/build

clean_macos_package:
	rm -f packaging/darwin/Distribution
	rm -f packaging/darwin/Resources/welcome.html
	rm -f packaging/darwin/scripts/postinstall
	rm -rf packaging/root/

clean_windows_msi:
	rm -rf packaging/windows/msi
	rm -f $(HOST_BUILD_DIR)/split

.PHONY: clean ## Remove all build artifacts
clean: clean_docs clean_macos_package clean_windows_msi
	rm -f $(GENERATED_RPM_FILES)
	rm -rf $(BUILD_DIR)
	rm -f $(GOPATH)/bin/crc
	rm -rf $(RELEASE_DIR)

.PHONY: build_e2e
build_e2e: $(SOURCES)
	GOOS=linux   go test ./test/e2e/ -c -o $(BUILD_DIR)/linux-amd64/e2e.test
	GOOS=windows go test ./test/e2e/ -c -o $(BUILD_DIR)/windows-amd64/e2e.test.exe
	GOOS=darwin  go test ./test/e2e/ -c -o $(BUILD_DIR)/macos-amd64/e2e.test

.PHONY: build_integration
build_integration: $(SOURCES)
	GOOS=linux   go test ./test/integration/ -c -o $(BUILD_DIR)/linux-amd64/integration.test
	GOOS=windows go test --ldflags="-X $(REPOPATH)/pkg/crc/version.installerBuild=true" ./test/integration/ -c -o $(BUILD_DIR)/windows-amd64/integration.test.exe
	GOOS=darwin  go test --ldflags="-X $(REPOPATH)/pkg/crc/version.installerBuild=true" ./test/integration/ -c -o $(BUILD_DIR)/macos-amd64/integration.test

#  Build the container image for e2e
.PHONY: containerized_e2e
containerized_e2e:
ifndef CRC_E2E_IMG_VERSION
CRC_E2E_IMG_VERSION=v$(CRC_VERSION)-$(COMMIT_SHA)
endif
IMG_E2E = quay.io/crcont/crc-e2e:$(CRC_E2E_IMG_VERSION)
containerized_e2e: clean
	$(CONTAINER_RUNTIME) build -t $(IMG_E2E) -f images/build-e2e/Dockerfile .

#  Build the container image for integration
.PHONY: containerized_integration
containerized_integration:
ifndef CRC_INTEGRATION_IMG_VERSION
CRC_INTEGRATION_IMG_VERSION=v$(CRC_VERSION)-$(COMMIT_SHA)
endif
IMG_INTEGRATION = quay.io/crcont/crc-integration:$(CRC_INTEGRATION_IMG_VERSION)
containerized_integration: clean
	$(CONTAINER_RUNTIME) build -t $(IMG_INTEGRATION) -f images/build-integration/Dockerfile .

.PHONY: integration ## Run integration tests in Ginkgo
integration:
ifndef PULL_SECRET_PATH
export PULL_SECRET_PATH = $(HOME)/Downloads/crc-pull-secret
endif
ifndef BUNDLE_PATH
export BUNDLE_PATH = $(HOME)/Downloads/crc_libvirt_$(OPENSHIFT_VERSION).$(BUNDLE_EXTENSION)
endif
integration:
	@go test -timeout=60m $(REPOPATH)/test/integration -v

.PHONY: e2e ## Run e2e tests
e2e:
GODOG_OPTS = --godog.tags=$(GOOS)
ifndef PULL_SECRET_FILE
	PULL_SECRET_FILE = --pull-secret-file=$(HOME)/Downloads/crc-pull-secret
endif
ifndef BUNDLE_LOCATION
	BUNDLE_LOCATION = --bundle-location=$(HOME)/Downloads/crc_libvirt_$(OPENSHIFT_VERSION).$(BUNDLE_EXTENSION)
endif
ifndef CRC_BINARY
	CRC_BINARY = --crc-binary=$(GOPATH)/bin
endif
e2e:
	@go test --timeout=180m $(REPOPATH)/test/e2e -v $(PULL_SECRET_FILE) $(BUNDLE_LOCATION) $(CRC_BINARY) $(GODOG_OPTS) $(CLEANUP_HOME) $(INSTALLER_PATH) $(USER_PASSWORD)

.PHONY: e2e-stories e2e-story-health e2e-story-marketplace e2e-story-registry
# cluster must already be running, crc must be in the path
e2e-stories: install e2e-story-health e2e-story-marketplace e2e-story-registry

e2e-story-health: install
	@go test $(REPOPATH)/test/e2e -v --godog.tags="$(GOOS) && ~@startstop && @story_health" --cleanup-home=false
e2e-story-marketplace: install
	@go test $(REPOPATH)/test/e2e -v --godog.tags="$(GOOS) && ~@startstop && @story_marketplace" --cleanup-home=false
e2e-story-registry: install
	@go test $(REPOPATH)/test/e2e -v --godog.tags="$(GOOS) && ~@startstop && @story_registry" --cleanup-home=false


.PHONY: fmt
fmt:
	@gofmt -l -w $(SOURCE_DIRS)

.PHONY: golangci-lint
golangci-lint:
	@if $(GOPATH)/bin/golangci-lint version 2>&1 | grep -vq $(GOLANGCI_LINT_VERSION); then\
		go install -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

# Run golangci-lint against code
.PHONY: lint cross-lint
lint: golangci-lint
	$(GOPATH)/bin/golangci-lint run

cross-lint: golangci-lint
	GOOS=darwin $(GOPATH)/bin/golangci-lint run
	GOOS=linux $(GOPATH)/bin/golangci-lint run
	GOOS=windows $(GOPATH)/bin/golangci-lint run

.PHONY: gen_release_info
gen_release_info:
	@cat release-info.json.sample | sed s/@CRC_VERSION@/$(CRC_VERSION)/ > $(RELEASE_INFO)
	@sed -i s/@GIT_COMMIT_SHA@/$(COMMIT_SHA)/ $(RELEASE_INFO)
	@sed -i s/@OPENSHIFT_VERSION@/$(OPENSHIFT_VERSION)/ $(RELEASE_INFO)
	@sed -i s/@PODMAN_VERSION@/$(PODMAN_VERSION)/ $(RELEASE_INFO)

.PHONY: linux-release-binary macos-release-binary windows-release-binary
linux-release-binary: LDFLAGS += -X '$(REPOPATH)/pkg/crc/version.linuxReleaseBuild=true' $(RELEASE_VERSION_VARIABLES)
linux-release-binary: $(BUILD_DIR)/linux-amd64/crc

macos-release-binary: LDFLAGS+= -X '$(REPOPATH)/pkg/crc/version.installerBuild=true' $(RELEASE_VERSION_VARIABLES)
macos-release-binary: $(BUILD_DIR)/macos-amd64/crc

windows-release-binary: LDFLAGS+= -X '$(REPOPATH)/pkg/crc/version.installerBuild=true' $(RELEASE_VERSION_VARIABLES)
windows-release-binary: $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: release linux-release
release: clean linux-release macos-release-binary windows-release-binary check
linux-release: clean lint linux-release-binary embed_crc_helpers gen_release_info
	mkdir $(RELEASE_DIR)

	@mkdir -p $(BUILD_DIR)/crc-linux-$(CRC_VERSION)-amd64
	@cp LICENSE $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/crc-linux-$(CRC_VERSION)-amd64
	tar cJSf $(RELEASE_DIR)/crc-linux-amd64.tar.xz -C $(BUILD_DIR) crc-linux-$(CRC_VERSION)-amd64 --owner=0 --group=0
	
	@mv $(RELEASE_INFO) $(RELEASE_DIR)/$(RELEASE_INFO)
	
	cd $(RELEASE_DIR) && sha256sum * > sha256sum.txt

.PHONY: embed_crc_helpers
embed_crc_helpers: $(BUILD_DIR)/linux-amd64/crc $(HOST_BUILD_DIR)/crc-embedder
ifeq ($(CUSTOM_EMBED),false)
	$(HOST_BUILD_DIR)/crc-embedder embed --log-level debug --goos=linux $(BUILD_DIR)/linux-amd64/crc
else
	$(HOST_BUILD_DIR)/crc-embedder embed --log-level debug --cache-dir=$(EMBED_DOWNLOAD_DIR) --no-download --goos=linux $(BUILD_DIR)/linux-amd64/crc
endif

.PHONY: update-go-version
update-go-version:
	./update-go-version.sh 1.16

.PHONY: goversioncheck
goversioncheck:
	./verify-go-version.sh

TRAY_RELEASE ?= packaging/tmp/crc-tray-macos.tar.gz

embed-download: $(HOST_BUILD_DIR)/crc-embedder
ifeq ($(CUSTOM_EMBED),false)
	mkdir -p $(EMBED_DOWNLOAD_DIR)
	$(HOST_BUILD_DIR)/crc-embedder download $(EMBED_DOWNLOAD_DIR)
endif

packagedir: clean embed-download macos-release-binary
	echo -n $(CRC_VERSION) > packaging/VERSION
	sed -e 's/__VERSION__/'$(CRC_VERSION)'/g' -e 's@__INSTALL_PATH__@$(MACOS_INSTALL_PATH)@g' packaging/darwin/Distribution.in >packaging/darwin/Distribution
	sed -e 's/__VERSION__/'$(CRC_VERSION)'/g' -e 's@__INSTALL_PATH__@$(MACOS_INSTALL_PATH)@g' packaging/darwin/welcome.html.in >packaging/darwin/Resources/welcome.html
	sed -e 's/__VERSION__/'$(CRC_VERSION)'/g' -e 's@__INSTALL_PATH__@$(MACOS_INSTALL_PATH)@g' packaging/darwin/postinstall.in >packaging/darwin/scripts/postinstall
	chmod 755 packaging/darwin/scripts/postinstall
	mkdir -p packaging/tmp/
	cp $(EMBED_DOWNLOAD_DIR)/* packaging/tmp/
	mkdir -p packaging/root/Applications
	tar -C packaging/root/Applications -xvzf $(TRAY_RELEASE)
	rm packaging/tmp/crc-tray-macos.tar.gz
	mv packaging/root/Applications/crc-tray-darwin-x64/crc-tray.app packaging/root/Applications/CodeReady\ Containers.app
	rm -fr packaging/root/Applications/crc-tray-darwin-x64

	mv packaging/tmp/* packaging/root/"$(MACOS_INSTALL_PATH)"

	cp $(BUILD_DIR)/macos-amd64/crc packaging/root/"$(MACOS_INSTALL_PATH)"
	cp LICENSE packaging/darwin/Resources/LICENSE.txt
	pkgbuild --analyze --root packaging/root packaging/components.plist
	plutil -replace BundleIsRelocatable -bool NO packaging/components.plist

$(BUILD_DIR)/macos-amd64/crc-macos-amd64.pkg: packagedir
	./packaging/package.sh $(BUILD_DIR)/macos-amd64

$(BUILD_DIR)/macos-amd64/crc-macos-installer-amd64.tar: packagedir
	tar -cvf $@ ./packaging
	cd $(@D) && sha256sum $(@F)>$(@F).sha256sum

$(GOPATH)/bin/gomod2rpmdeps:
	pushd /tmp && GO111MODULE=on go get github.com/cfergeau/gomod2rpmdeps/cmd/gomod2rpmdeps && popd

%.spec: %.spec.in $(GOPATH)/bin/gomod2rpmdeps
	@$(GOPATH)/bin/gomod2rpmdeps | sed -e '/__BUNDLED_PROVIDES__/r /dev/stdin' \
					   -e '/__BUNDLED_PROVIDES__/d' \
					   -e 's/__VERSION__/'$(CRC_VERSION)'/g' \
					   -e 's/__OPENSHIFT_VERSION__/'$(OPENSHIFT_VERSION)'/g' \
				       $< >$@

%: %.in
	@sed -e 's/__VERSION__/'$(CRC_VERSION)'/g' \
	     -e 's/__OPENSHIFT_VERSION__/'$(OPENSHIFT_VERSION)'/g' \
	     $< >$@

$(HOST_BUILD_DIR)/GenMsiWxs: packaging/windows/gen_msi_wxs.go
	go build -o $@ -ldflags="-X main.crcVersion=$(CRC_VERSION)" $<

CRC_EXE=crc.exe
BUNDLE_NAME=crc_hyperv_$(OPENSHIFT_VERSION).$(BUNDLE_EXTENSION)

.PHONY: msidir
msidir: clean embed-download $(HOST_BUILD_DIR)/GenMsiWxs windows-release-binary $(PACKAGE_DIR)/product.wxs.template
	mkdir -p $(PACKAGE_DIR)/msi
	cp $(EMBED_DOWNLOAD_DIR)/* $(PACKAGE_DIR)/msi
	cp $(HOST_BUILD_DIR)/crc.exe $(PACKAGE_DIR)/msi/$(CRC_EXE)
	pwsh -NoProfile -Command "cd $(PACKAGE_DIR)/msi; Expand-Archive crc-tray-windows.zip -DestinationPath .\; Remove-Item crc-tray-windows.zip"
	$(HOST_BUILD_DIR)/GenMsiWxs
	cp -r $(PACKAGE_DIR)/Resources $(PACKAGE_DIR)/msi/
	cp $(PACKAGE_DIR)/*.wxs $(PACKAGE_DIR)/msi
	rm $(PACKAGE_DIR)/product.wxs

$(BUILD_DIR)/windows-amd64/crc-windows-amd64.msi: msidir
	heat.exe dir $(PACKAGE_DIR)/msi/crc-tray-win32-x64 -o $(PACKAGE_DIR)/msi/trayfiles.wxs -cg CrcTrayFiles -scom -sfrag -srd -sreg -gg -g1 -dr TRAYDIR
	candle.exe -arch x64 -ext WixUtilExtension -o $(PACKAGE_DIR)/msi/ $(PACKAGE_DIR)/msi/*.wxs
	light.exe -ext WixUIExtension -ext WixUtilExtension -sacl -spdb -sice:ICE61 -sice:ICE69 -b $(PACKAGE_DIR)/msi -b $(PACKAGE_DIR)/msi/crc-tray-win32-x64 -out $@ $(PACKAGE_DIR)/msi/*.wixobj

CABS_MSI = "*.cab,crc-windows-amd64.msi"
$(BUILD_DIR)/windows-amd64/crc-windows-installer.zip: $(BUILD_DIR)/windows-amd64/crc-windows-amd64.msi
	rm -f $(HOST_BUILD_DIR)/crc.exe
	rm -f $(HOST_BUILD_DIR)/crc-embedder
	rm -f $(HOST_BUILD_DIR)/split
	pwsh -NoProfile -Command "cd $(HOST_BUILD_DIR); Compress-Archive -Path $(CABS_MSI) -DestinationPath crc-windows-installer.zip"
	cd $(@D) && sha256sum $(@F)>$(@F).sha256sum
