all: install

SHELL := /bin/bash -o pipefail

OPENSHIFT_VERSION ?= 4.19.3
OKD_VERSION ?= 4.19.0-okd-scos.1
MICROSHIFT_VERSION ?= 4.19.0
BUNDLE_EXTENSION = crcbundle
CRC_VERSION = 2.53.0
COMMIT_SHA?=$(shell git rev-parse --short=6 HEAD)
MACOS_INSTALL_PATH = /usr/local/crc
CONTAINER_RUNTIME ?= podman

TOOLS_DIR := tools
include tools/tools.mk

# Go and compilation related variables
BUILD_DIR ?= out
SOURCE_DIRS = cmd pkg test
RELEASE_DIR ?= release

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOVERSION = 1.23

HOST_BUILD_DIR=$(BUILD_DIR)/$(GOOS)-$(GOARCH)
GOPATH ?= $(shell go env GOPATH)
ORG := github.com/crc-org
MODULEPATH = $(ORG)/crc/v2
PACKAGE_DIR := packaging/$(GOOS)

SOURCES := $(shell git ls-files '*.go' ":^vendor")
SOURCES := $(SOURCES) go.mod go.sum Makefile

RELEASE_INFO := release-info.json

CUSTOM_EMBED ?= false
EMBED_DOWNLOAD_DIR ?= tmp-embed

SELINUX_VOLUME_LABEL = :Z
ifeq ($(GOOS),darwin)
SELINUX_VOLUME_LABEL :=
endif

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
VERSION_VARIABLES := -X $(MODULEPATH)/pkg/crc/version.crcVersion=$(CRC_VERSION) \
	-X $(MODULEPATH)/pkg/crc/version.ocpVersion=$(OPENSHIFT_VERSION) \
	-X $(MODULEPATH)/pkg/crc/version.okdVersion=$(OKD_VERSION) \
	-X $(MODULEPATH)/pkg/crc/version.microshiftVersion=$(MICROSHIFT_VERSION) \
	-X $(MODULEPATH)/pkg/crc/version.commitSha=$(COMMIT_SHA)
RELEASE_VERSION_VARIABLES := -X $(MODULEPATH)/pkg/crc/segment.WriteKey=cvpHsNcmGCJqVzf6YxrSnVlwFSAZaYtp

# https://golang.org/cmd/link/
LDFLAGS := $(VERSION_VARIABLES) ${GO_EXTRA_LDFLAGS}
# Same build flags are used in the podman remote to cross build it https://github.com/containers/podman/blob/main/Makefile
BUILDTAGS := containers_image_openpgp

# Add default target
.PHONY: default
default: install

# Create and update the vendor directory
.PHONY: vendor
vendor:
	go mod tidy -go $(GOVERSION).0
	go mod vendor

.PHONY: vendorcheck
vendorcheck:
	./verify-vendor.sh

.PHONY: check
check: cross build_e2e_all $(HOST_BUILD_DIR)/crc-embedder test cross-lint vendorcheck build_integration_all

# Start of the actual build targets

.PHONY: install
install: $(SOURCES)
	go install -tags "$(BUILDTAGS)"  -ldflags="$(LDFLAGS)" $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(BUILD_DIR)/macos-amd64/crc: $(SOURCES)
	GOARCH=amd64 GOOS=darwin go build -tags "$(BUILDTAGS)" -ldflags="$(LDFLAGS)" -o $@ $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(BUILD_DIR)/macos-arm64/crc: $(SOURCES)
	GOARCH=arm64 GOOS=darwin go build -tags "$(BUILDTAGS)" -ldflags="$(LDFLAGS)" -o $@ $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(BUILD_DIR)/linux-amd64/crc: $(SOURCES)
	GOOS=linux GOARCH=amd64 go build -tags "$(BUILDTAGS)" -ldflags="$(LDFLAGS)" -o $@ $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(BUILD_DIR)/linux-arm64/crc: $(SOURCES)
	GOOS=linux GOARCH=arm64 go build -tags "$(BUILDTAGS)" -ldflags="$(LDFLAGS)" -o $@ $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(BUILD_DIR)/windows-amd64/crc.exe: $(SOURCES)
	GOARCH=amd64 GOOS=windows go build -tags "$(BUILDTAGS)" -ldflags="$(LDFLAGS)" -o $@ $(GO_EXTRA_BUILDFLAGS) ./cmd/crc

$(HOST_BUILD_DIR)/crc-embedder: $(SOURCES)
	go build --tags="build" -ldflags="$(LDFLAGS)" -o $(HOST_BUILD_DIR)/crc-embedder $(GO_EXTRA_BUILDFLAGS) ./cmd/crc-embedder

.PHONY: cross ## Cross compiles all binaries
cross: $(BUILD_DIR)/macos-arm64/crc $(BUILD_DIR)/macos-amd64/crc $(BUILD_DIR)/linux-amd64/crc $(BUILD_DIR)/linux-arm64/crc $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: containerized ## Cross compile from container
containerized: clean
	${CONTAINER_RUNTIME} build -t crc-build -f images/build .
	${CONTAINER_RUNTIME} run --name crc-cross crc-build make cross
	${CONTAINER_RUNTIME} cp crc-cross:/opt/app-root/src/out ./
	${CONTAINER_RUNTIME} rm crc-cross
	${CONTAINER_RUNTIME} rmi crc-build

.PHONY: generate_mocks
generate_mocks: $(TOOLS_BINDIR)/mockery
	$(TOOLS_BINDIR)/mockery --srcpkg ./pkg/crc/api/client --name Client --output test/mocks/api  --filename client.go

.PHONY: test
test: gen_release_info
	go test -race --tags "build $(BUILDTAGS)" -v -ldflags="$(VERSION_VARIABLES)" . ./pkg/... ./cmd/...

.PHONY: spec test-rpmbuild

GENERATED_RPM_FILES=packaging/rpm/crc.spec images/rpmbuild/Containerfile
spec: $(GENERATED_RPM_FILES)

test-rpmbuild: spec
	${CONTAINER_RUNTIME} build -t test-rpmbuild-img -f images/rpmbuild/Containerfile .
	${CONTAINER_RUNTIME} create --name test-rpmbuild test-rpmbuild-img
	${CONTAINER_RUNTIME} cp test-rpmbuild:/root/rpmbuild/RPMS/ .
	${CONTAINER_RUNTIME} cp test-rpmbuild:/root/rpmbuild/BUILD/crc-$(CRC_VERSION)-$(OPENSHIFT_VERSION)/out/linux-amd64/crc .
	${CONTAINER_RUNTIME} rm test-rpmbuild
	${CONTAINER_RUNTIME} rmi test-rpmbuild-img

clean_macos_package:
	rm -f packaging/darwin/Distribution
	rm -f packaging/darwin/Resources/welcome.html
	rm -f packaging/darwin/scripts/postinstall
	rm -rf packaging/darwin/root-crc/

clean_windows_msi:
	rm -rf packaging/windows/msi
	rm -f $(HOST_BUILD_DIR)/split

.PHONY: clean ## Remove all build artifacts
clean: clean_macos_package clean_windows_msi
	rm -f $(GENERATED_RPM_FILES)
	rm -rf $(BUILD_DIR)
	rm -f $(GOPATH)/bin/crc
	rm -rf $(RELEASE_DIR)

## e2e building section
.PHONY: build_e2e build_e2e_all containerized_e2e

## Function to build the e2e binary params: (os, param, binary_name)
e2e_builder = GOOS=$(1) GOARCH=$(2) go test ./test/e2e/ -tags "$(BUILDTAGS)" --ldflags="$(VERSION_VARIABLES)" -c -o $(BUILD_DIR)/$(1)-$(2)/$(3)

build_e2e: $(SOURCES)
	$(call e2e_builder,$(GOOS),$(GOARCH),e2e.test)

build_e2e_all: $(SOURCES) 
	$(call e2e_builder,linux,amd64,e2e.test)
	$(call e2e_builder,windows,amd64,e2e.test.exe)
	$(call e2e_builder,darwin,amd64,e2e.test)
	$(call e2e_builder,darwin,arm64,e2e.test)

containerized_e2e:
ifndef CRC_E2E_IMG_VERSION
CRC_E2E_IMG_VERSION=v$(CRC_VERSION)-$(COMMIT_SHA)
endif
IMG_E2E = quay.io/crcont/crc-e2e:$(CRC_E2E_IMG_VERSION)
containerized_e2e: clean
	${CONTAINER_RUNTIME} build -t ${IMG_E2E}-${OS}-${ARCH} -f images/build-e2e/Containerfile --build-arg=OS=${OS} --build-arg=ARCH=${ARCH} .

## integration building section
.PHONY: build_integration build_integration_all containerized_integration

## Function to build the integration binary params: (os, param, ldflags, binary_name)
integration_builder = GOOS=$(1) GOARCH=$(2) go test ./test/integration/ -tags "$(BUILDTAGS)" --ldflags="$(3)" -c -o $(BUILD_DIR)/$(1)-$(2)/$(4)

build_integration: 
ILDFLAGS=$(LDFLAGS)
ifneq ($(GOOS), linux)
ILDFLAGS=$(LDFLAGS) -X $(MODULEPATH)/pkg/crc/version.installerBuild=true
endif
build_integration: $(SOURCES)
	$(call integration_builder,$(GOOS),$(GOARCH),$(ILDFLAGS),integration.test)

build_integration_all: $(SOURCES) 
	$(call integration_builder,linux,amd64,$(LDFLAGS),integration.test)
	$(call integration_builder,windows,amd64,$(LDFLAGS) -X $(MODULEPATH)/pkg/crc/version.installerBuild=true,integration.test.exe)
	$(call integration_builder,darwin,amd64,$(LDFLAGS) -X $(MODULEPATH)/pkg/crc/version.installerBuild=true,integration.test)
	$(call integration_builder,darwin,arm64,$(LDFLAGS) -X $(MODULEPATH)/pkg/crc/version.installerBuild=true,integration.test)

containerized_integration: 
ifndef CRC_INTEGRATION_IMG_VERSION
CRC_INTEGRATION_IMG_VERSION=v$(CRC_VERSION)-$(COMMIT_SHA)
endif
IMG_INTEGRATION = quay.io/crcont/crc-integration:$(CRC_INTEGRATION_IMG_VERSION)
containerized_integration: clean
	$(CONTAINER_RUNTIME) build -t $(IMG_INTEGRATION)-${OS}-${ARCH} -f images/build-integration/Containerfile --build-arg=OS=${OS} --build-arg=ARCH=${ARCH} .

.PHONY: integration ## Run integration tests in Ginkgo
integration:
ifndef GINKGO_OPTS
GINKGO_OPTS = --ginkgo.label-filter=""
endif
ifndef PULL_SECRET_PATH
PULL_SECRET_PATH = --pull-secret-path=$(HOME)/Downloads/crc-pull-secret
endif
ifndef BUNDLE_PATH
BUNDLE_PATH = --bundle-path=$(HOME)/Downloads/crc_libvirt_$(OPENSHIFT_VERSION)_$(GOARCH).$(BUNDLE_EXTENSION)
endif

integration:
	@go test -timeout=90m -tags "$(BUILDTAGS)" $(MODULEPATH)/test/integration $(PULL_SECRET_PATH) $(BUNDLE_PATH) -v $(GINKGO_OPTS) 

.PHONY: e2e ## Run e2e tests
e2e:
GODOG_OPTS = --godog.tags=$(GOOS)
ifndef PULL_SECRET_FILE
	PULL_SECRET_FILE = --pull-secret-file=$(HOME)/Downloads/crc-pull-secret
endif
ifndef BUNDLE_LOCATION
	BUNDLE_LOCATION = --bundle-location=$(HOME)/Downloads/crc_libvirt_$(OPENSHIFT_VERSION)_$(GOARCH).$(BUNDLE_EXTENSION)
endif
ifndef CRC_BINARY
	CRC_BINARY = --crc-binary=$(GOPATH)/bin
endif
e2e:
	@go test --timeout=180m $(MODULEPATH)/test/e2e -tags "$(BUILDTAGS)" --ldflags="$(VERSION_VARIABLES)" -v $(PULL_SECRET_FILE) $(BUNDLE_LOCATION) $(CRC_BINARY) $(GODOG_OPTS) $(CLEANUP_HOME) $(VERSION_TO_TEST)

.PHONY: e2e-stories e2e-story-health e2e-story-marketplace e2e-story-registry
# cluster must already be running, crc must be in the path
e2e-stories: install e2e-story-health e2e-story-marketplace e2e-story-registry

e2e-story-health: install
	@go test $(MODULEPATH)/test/e2e --ldflags="$(VERSION_VARIABLES)" -v $(CRC_BINARY) --godog.tags="$(GOOS) && ~@startstop && @story_health" --cleanup-home=false
e2e-story-marketplace: install
	@go test $(MODULEPATH)/test/e2e --ldflags="$(VERSION_VARIABLES)" -v $(CRC_BINARY) --godog.tags="$(GOOS) && ~@startstop && @story_marketplace" --cleanup-home=false
e2e-story-registry: install
	@go test $(MODULEPATH)/test/e2e --ldflags="$(VERSION_VARIABLES)" -v $(CRC_BINARY) --godog.tags="$(GOOS) && ~@startstop && @story_registry" --cleanup-home=false
e2e-story-microshift: install
	@go test $(MODULEPATH)/test/e2e -tags "$(BUILDTAGS)" --ldflags="$(VERSION_VARIABLES)" -v $(PULL_SECRET_FILE) $(BUNDLE_LOCATION) $(CRC_BINARY) --godog.tags="$(GOOS) && @microshift" --cleanup-home=false

.PHONY: fmt
fmt: $(TOOLS_BINDIR)/goimports
	@$(TOOLS_BINDIR)/goimports -l -w $(SOURCE_DIRS)

# Run golangci-lint against code
.PHONY: lint cross-lint
lint: $(TOOLS_BINDIR)/golangci-lint gen_release_info
	"$(TOOLS_BINDIR)"/golangci-lint run

cross-lint: $(TOOLS_BINDIR)/golangci-lint gen_release_info
	GOARCH=amd64 GOOS=darwin "$(TOOLS_BINDIR)"/golangci-lint run
	GOARCH=arm64 GOOS=darwin "$(TOOLS_BINDIR)"/golangci-lint run
	GOARCH=amd64 GOOS=linux "$(TOOLS_BINDIR)"/golangci-lint run
	GOARCH=amd64 GOOS=windows "$(TOOLS_BINDIR)"/golangci-lint run

.PHONY: gen_release_info
gen_release_info:
	@cat release-info.json.sample | sed s/@CRC_VERSION@/$(CRC_VERSION)/ > $(RELEASE_INFO)
	@sed -i.bak -e s/@GIT_COMMIT_SHA@/$(COMMIT_SHA)/ $(RELEASE_INFO)
	@sed -i.bak -e s/@OPENSHIFT_VERSION@/$(OPENSHIFT_VERSION)/ $(RELEASE_INFO)
	@rm -f $(RELEASE_INFO).bak

.PHONY: linux-release-binary macos-release-binary windows-release-binary
linux-release-binary: LDFLAGS+= $(RELEASE_VERSION_VARIABLES)
linux-release-binary: $(BUILD_DIR)/linux-${GOARCH}/crc

macos-release-binary: LDFLAGS+= -X '$(MODULEPATH)/pkg/crc/version.installerBuild=true' $(RELEASE_VERSION_VARIABLES)
macos-release-binary: $(BUILD_DIR)/macos-universal/crc

windows-release-binary: LDFLAGS+= -X '$(MODULEPATH)/pkg/crc/version.installerBuild=true' $(RELEASE_VERSION_VARIABLES)
windows-release-binary: $(BUILD_DIR)/windows-amd64/crc.exe

.PHONY: release linux-release
release: clean linux-release macos-release-binary windows-release-binary check
linux-release: clean lint linux-release-binary embed_crc_helpers gen_release_info
	mkdir $(RELEASE_DIR)

	@mkdir -p $(BUILD_DIR)/crc-linux-$(CRC_VERSION)-${GOARCH}
	@cp LICENSE $(BUILD_DIR)/linux-${GOARCH}/crc $(BUILD_DIR)/crc-linux-$(CRC_VERSION)-${GOARCH}
	tar cJSf $(RELEASE_DIR)/crc-linux-${GOARCH}.tar.xz -C $(BUILD_DIR) crc-linux-$(CRC_VERSION)-${GOARCH} --owner=0 --group=0

	@cp $(RELEASE_INFO) $(RELEASE_DIR)/$(RELEASE_INFO)

	cd $(RELEASE_DIR) && sha256sum * > sha256sum.txt

.PHONY: embed_crc_helpers
embed_crc_helpers: $(BUILD_DIR)/linux-${GOARCH}/crc $(HOST_BUILD_DIR)/crc-embedder
ifeq ($(CUSTOM_EMBED),false)
	$(HOST_BUILD_DIR)/crc-embedder embed --log-level debug --goos=linux $(BUILD_DIR)/linux-${GOARCH}/crc
else
	$(HOST_BUILD_DIR)/crc-embedder embed --log-level debug --cache-dir=$(EMBED_DOWNLOAD_DIR) --no-download --goos=linux $(BUILD_DIR)/linux-${GOARCH}/crc
endif

.PHONY: update-go-version
update-go-version:
	./update-go-version.sh $(GOVERSION)

.PHONY: goversioncheck
goversioncheck:
	./verify-go-version.sh

.PHONY: embed-download-windows embed-download-darwin
embed-download-windows embed-download-darwin: embed-download-%: $(HOST_BUILD_DIR)/crc-embedder
ifeq ($(CUSTOM_EMBED),false)
	mkdir -p $(EMBED_DOWNLOAD_DIR)
	$(HOST_BUILD_DIR)/crc-embedder download --goos=$* $(EMBED_DOWNLOAD_DIR)
endif

$(BUILD_DIR)/macos-universal/crc: $(BUILD_DIR)/macos-arm64/crc $(BUILD_DIR)/macos-amd64/crc $(TOOLS_BINDIR)/makefat
	mkdir -p out/macos-universal
	cd $(BUILD_DIR) && "$(TOOLS_BINDIR)"/makefat macos-universal/crc macos-amd64/crc macos-arm64/crc

packagedir: clean_macos_package embed-download-darwin macos-release-binary
	echo -n $(CRC_VERSION) > packaging/darwin/VERSION

	mkdir -p packaging/darwin/root-crc/Applications

	# crc.pkg
	#ls $(EMBED_DOWNLOAD_DIR)
	mkdir -p packaging/darwin/root-crc/"$(MACOS_INSTALL_PATH)"
	mv $(EMBED_DOWNLOAD_DIR)/vf.entitlements packaging/darwin/vfkit.entitlements
	mv $(EMBED_DOWNLOAD_DIR)/* packaging/darwin/root-crc/"$(MACOS_INSTALL_PATH)"
	cp $(BUILD_DIR)/macos-universal/crc packaging/darwin/root-crc/"$(MACOS_INSTALL_PATH)"

	# Resources used by `productbuild`
	sed -e 's/__VERSION__/'$(CRC_VERSION)'/g' -e 's@__INSTALL_PATH__@$(MACOS_INSTALL_PATH)@g' packaging/darwin/Distribution.in >packaging/darwin/Distribution
	sed -e 's/__VERSION__/'$(CRC_VERSION)'/g' -e 's@__INSTALL_PATH__@$(MACOS_INSTALL_PATH)@g' packaging/darwin/welcome.html.in >packaging/darwin/Resources/welcome.html
	sed -e 's/__VERSION__/'$(CRC_VERSION)'/g' -e 's@__INSTALL_PATH__@$(MACOS_INSTALL_PATH)@g' packaging/darwin/postinstall.in >packaging/darwin/scripts/postinstall
	chmod 755 packaging/darwin/scripts/postinstall
	cp LICENSE packaging/darwin/Resources/LICENSE.txt

$(BUILD_DIR)/macos-universal/crc-macos-installer.pkg: packagedir
	./packaging/darwin/macos-pkg-build-and-sign.sh $(@D)

$(BUILD_DIR)/macos-universal/crc-macos-installer.tar: packagedir
	tar -C ./packaging -cvf $@ darwin
	cd $(@D) && sha256sum $(@F)>$(@F).sha256sum

%.spec: %.spec.in $(TOOLS_BINDIR)/gomod2rpmdeps
	@"$(TOOLS_BINDIR)"/gomod2rpmdeps | sed -e '/__BUNDLED_PROVIDES__/r /dev/stdin' \
					   -e '/__BUNDLED_PROVIDES__/d' \
					   -e 's/__VERSION__/'$(CRC_VERSION)'/g' \
					   -e 's/__OPENSHIFT_VERSION__/'$(OPENSHIFT_VERSION)'/g' \
					   -e 's/__COMMIT_SHA__/'$(COMMIT_SHA)'/g' \
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
msidir: clean_windows_msi embed-download-windows $(HOST_BUILD_DIR)/GenMsiWxs windows-release-binary $(PACKAGE_DIR)/product.wxs.template
	mkdir -p $(PACKAGE_DIR)/msi
	cp $(EMBED_DOWNLOAD_DIR)/* $(PACKAGE_DIR)/msi
	cp $(HOST_BUILD_DIR)/crc.exe $(PACKAGE_DIR)/msi/$(CRC_EXE)
	$(HOST_BUILD_DIR)/GenMsiWxs
	cp -r $(PACKAGE_DIR)/Resources $(PACKAGE_DIR)/msi/
	cp $(PACKAGE_DIR)/*.wxs $(PACKAGE_DIR)/msi
	rm $(PACKAGE_DIR)/product.wxs

$(BUILD_DIR)/windows-amd64/crc-windows-amd64.msi: msidir
	candle.exe -arch x64 -ext WixUtilExtension -o $(PACKAGE_DIR)/msi/ $(PACKAGE_DIR)/msi/*.wxs
	light.exe -ext WixUIExtension -ext WixUtilExtension -sacl -spdb -sice:ICE61 -sice:ICE69 -b $(PACKAGE_DIR)/msi -loc $(PACKAGE_DIR)/WixUI_en.wxl -out $@ $(PACKAGE_DIR)/msi/*.wixobj

CABS_MSI = "*.cab,crc-windows-amd64.msi"
$(BUILD_DIR)/windows-amd64/crc-windows-installer.zip: $(BUILD_DIR)/windows-amd64/crc-windows-amd64.msi
	rm -f $(HOST_BUILD_DIR)/crc.exe
	rm -f $(HOST_BUILD_DIR)/crc-embedder
	rm -f $(HOST_BUILD_DIR)/split
	pwsh -NoProfile -Command "cd $(HOST_BUILD_DIR); Compress-Archive -Path $(CABS_MSI) -DestinationPath crc-windows-installer.zip"
	cd $(@D) && sha256sum $(@F)>$(@F).sha256sum

.PHONY: choco choco-clean
CHOCO_PKG_DIR = packaging/chocolatey/crc
$(CHOCO_PKG_DIR)/tools/crc-admin-helper-windows.exe: $(HOST_BUILD_DIR)/crc-embedder
	$(HOST_BUILD_DIR)/crc-embedder download --goos=windows --components=admin-helper $(CHOCO_PKG_DIR)/tools
choco: clean choco-clean $(BUILD_DIR)/windows-amd64/crc.exe $(CHOCO_PKG_DIR)/tools/crc-admin-helper-windows.exe $(CHOCO_PKG_DIR)/crc.nuspec $(CHOCO_PKG_DIR)/VERIFICATION.txt
	cp $(BUILD_DIR)/windows-amd64/crc.exe $(CHOCO_PKG_DIR)/tools/crc.exe
	mv $(CHOCO_PKG_DIR)/VERIFICATION.txt $(CHOCO_PKG_DIR)/tools/VERIFICATION.txt
	powershell.exe -NoProfile -Command "@('From: https://github.com/crc-org/crc/blob/main/LICENSE') + (Get-Content 'LICENSE') | Set-Content $(CHOCO_PKG_DIR)/tools/LICENSE.txt"
	cd $(CHOCO_PKG_DIR) && choco pack
choco-clean:
	rm -f $(CHOCO_PKG_DIR)/*.nupkg
	rm -f $(CHOCO_PKG_DIR)/tools/*.exe
	rm -f $(CHOCO_PKG_DIR)/crc.nuspec
	rm -f $(CHOCO_PKG_DIR)/tools/VERIFICATION.txt

ADMIN_HELPER_HASH = $(shell powershell.exe -NoProfile -Command "Get-FileHash -Algorithm SHA256 $(CHOCO_PKG_DIR)/tools/crc-admin-helper-windows.exe | Select-Object -ExpandProperty Hash")
HELPER_SCRIPT_HASH = $(shell powershell.exe -NoProfile -Command "Get-FileHash -Algorithm SHA256 $(CHOCO_PKG_DIR)/tools/crcprerequisitesetup.ps1 | Select-Object -ExpandProperty Hash")
# todo: retreive this dynamically instead of setting here
ADMIN_HELPER_VERSION = 0.0.12
%.txt: %.txt.in
	@sed -e 's/__ADMIN_HELPER_CHECKSUM__/'$(ADMIN_HELPER_HASH)'/g' \
		 -e 's/__HELPER_SCRIPT_CHECKSUM__/'$(HELPER_SCRIPT_HASH)'/g' \
		 -e 's/__ADMIN_HELPER_VERSION__/'$(ADMIN_HELPER_VERSION)'/g' \
	     $< >$@
