TOOLS_BINDIR = $(realpath $(TOOLS_DIR)/bin)

$(TOOLS_BINDIR)/makefat: $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && GOBIN="$(TOOLS_BINDIR)" go install github.com/randall77/makefat

$(TOOLS_BINDIR)/golangci-lint: $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && GOBIN="$(TOOLS_BINDIR)" go install github.com/golangci/golangci-lint/cmd/golangci-lint

$(TOOLS_BINDIR)/gomod2rpmdeps: $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && GOBIN="$(TOOLS_BINDIR)" go install github.com/cfergeau/gomod2rpmdeps/cmd/gomod2rpmdeps

$(TOOLS_BINDIR)/mockery: $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && GOBIN="$(TOOLS_BINDIR)" go install github.com/vektra/mockery/v2@latest

$(TOOLS_BINDIR)/goimports: $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && GOBIN="$(TOOLS_BINDIR)" go install golang.org/x/tools/cmd/goimports@latest
