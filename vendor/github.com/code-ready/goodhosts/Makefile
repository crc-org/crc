# Go and compilation related variables
BUILD_DIR ?= out

ORG := github.com/code-ready
REPOPATH ?= $(ORG)/goodhosts

vendor:
	go mod vendor

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf vendor

.PHONY: build
build: $(BUILD_DIR) vendor
	go build \
		-installsuffix "static" \
		-o $(BUILD_DIR)/goodhosts \
                cmd/main.go
	chmod +x $(BUILD_DIR)/goodhosts
