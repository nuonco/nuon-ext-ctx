NUON_REPO_ROOT ?= /Users/harsh/work/nuonco/nuon
EXTENSION_PKG := ./bins/cli/extensions/nuon-ext-ctx
EXTENSION_PKGS := ./bins/cli/extensions/nuon-ext-ctx/...
BINARY := nuon-ext-ctx

.PHONY: build test fmt vet clean check-repo

check-repo:
	@test -f "$(NUON_REPO_ROOT)/go.mod" || \
		(echo "NUON_REPO_ROOT must point to the Nuon monorepo root (missing go.mod at $(NUON_REPO_ROOT))" && exit 1)

build: check-repo
	go -C "$(NUON_REPO_ROOT)" build -o "$(CURDIR)/$(BINARY)" "$(EXTENSION_PKG)"

test: check-repo
	go -C "$(NUON_REPO_ROOT)" test "$(EXTENSION_PKGS)"

fmt: check-repo
	go -C "$(NUON_REPO_ROOT)" fmt "$(EXTENSION_PKGS)"

vet: check-repo
	go -C "$(NUON_REPO_ROOT)" vet "$(EXTENSION_PKGS)"

clean:
	rm -f "$(BINARY)"
