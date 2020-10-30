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

PLATFORMS := darwin-amd64 linux-386 linux-amd64
#linux-arm linux-arm64 freebsd-386 freebsd-amd64 openbsd-386 openbsd-x64 solaris-amd64 windows-amd64

default: build test readme

build: $(DIST)/$(BIN_NAME)
$(DIST)/$(BIN_NAME):
	go build $(ARGS) -ldflags="$(LDFLAGS)" -o $@

all: $(addsuffix /$(BIN), $(addprefix $(DIST)/,$(PLATFORMS)))

clean:
	rm -rf dist && mkdir dist

$(DIST)/%/$(BIN_NAME): GOOS   = $(word 1,$(subst -, ,$*))
$(DIST)/%/$(BIN_NAME): GOARCH = $(word 2,$(subst -, ,$*))
$(DIST)/%/$(BIN_NAME): $(SOURCE)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(ARGS) -ldflags="$(LDFLAGS)" -o $@
	cd dist && shasum -a 256 hclq-* > hclq-shasums

shasums:
	cd "$(DIST)" && shasum -a 256 $(BIN_NAME)-* > hclq-shasums

install:
	#go install -mod=vendor -ldflags="${LDFLAGS}"
	go install -ldflags=$(LDFLAGS)

readme: README.md
README.md: README.md.rb
	#@if [ ! -f .git/hooks/pre-commit ]; then printf "Missing pre-commit hook for readme, be sure to copy it from hclq-pages repo"; exit 1; fi
	erb README.md.rb > README.md

test: $(GO_JUNIT_REPORT) build
	go test -v "./..."
	#@mkdir -p test


.PHONY: clean install test