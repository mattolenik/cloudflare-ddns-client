# Version and linker flags
# This will return either the current tag, branch, or commit hash of this repo.
REPO_NAME      := cloudflare-ddns
VERSION         = $(shell echo $$(ver=$$(git tag -l --points-at HEAD) && [ -z $$ver ] && ver=$$(git describe --always --dirty); printf $$ver))
LDFLAGS         = -s -w -X main.version=$(VERSION)
PROJECT_ROOT    = $(shell cd -P -- '$(shell dirname -- "$0")' && pwd -P)
BIN_NAME       := cloudflare-ddns
DIST            = dist

SOURCE := $(shell find $(PROJECT_ROOT) -name '*.go')
BINS   := $(shell find $(DIST) -name '$(BIN_NAME)-*')

PLATFORMS ?= darwin-amd64 dragonfly-amd64 freebsd-amd64 freebsd-arm freebsd-arm64 linux-amd64 linux-arm linux-arm64 netbsd-amd64 netbsd-arm netbsd-arm64 openbsd-amd64 openbsd-arm openbsd-arm64 windows-amd64 windows-arm

default: all shasums test readme

build: $(DIST)/$(BIN_NAME)
$(DIST)/$(BIN_NAME): $(SOURCE)
	go build $(ARGS) -ldflags="$(LDFLAGS)" -o $@

all: $(addprefix $(DIST)/$(BIN_NAME)-,$(PLATFORMS))

clean:
	rm -rf dist && mkdir dist

$(DIST)/$(BIN_NAME)-%: GOOS   = $(word 1,$(subst -, ,$*))
$(DIST)/$(BIN_NAME)-%: GOARCH = $(word 2,$(subst -, ,$*))
$(DIST)/$(BIN_NAME)-%: $(SOURCE)
	CGO=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(ARGS) -ldflags="$(LDFLAGS)" -o $@

shasums: $(DIST)/$(BIN_NAME)-shasums
$(DIST)/$(BIN_NAME)-shasums: all
	cd $(DIST) && shasum -a 256 $(BIN_NAME)-* > $(BIN_NAME)-shasums

install:
	#go install -mod=vendor -ldflags="${LDFLAGS}"
	go install -ldflags=$(LDFLAGS)

readme: README.md
README.md: README.tpl.md
	go run tools/readmegen/main.go README.tpl.md > README.md

test: $(GO_JUNIT_REPORT) build
	go test -v "./..."
	#@mkdir -p test


.PHONY: clean install test