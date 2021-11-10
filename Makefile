# Version and linker flags
# This will return either the current tag, branch, or commit hash of this repo.
MODULE               := $(shell awk 'NR==1{print $$2}' go.mod)
REPO_NAME            := cloudflare-ddns
VERSION              := $(shell echo $$(ver=$$(git tag -l --points-at HEAD) && [ -z $$ver ] && ver=$$(git describe --always --dirty); printf $$ver))
# This is set during CI in .github/workflows/release.yml
RELEASE_VERSION      ?= $(VERSION)
LDFLAGS              := -s -w -X $(MODULE)/conf.Version=$(VERSION) -X $(MODULE)/conf.ModuleName=$(MODULE)
FLAGS                := -trimpath
PROJECT_ROOT         := $(shell cd -P -- '$(shell dirname -- "$0")' && pwd -P)
BIN_NAME             := cloudflare-ddns
DIST                 := dist
UNAME                := $(shell uname)
GOTESTSUM            := go run gotest.tools/gotestsum
GOVERALLS            := go run github.com/mattn/goveralls
SOURCE               := $(shell find $(PROJECT_ROOT) -name '*.go')
BINS                 := $(shell find $(DIST) -name '$(BIN_NAME)-*')
PLATFORMS            ?= darwin-amd64 dragonfly-amd64 freebsd-amd64 freebsd-arm freebsd-arm64 linux-amd64 linux-arm linux-arm64 netbsd-amd64 netbsd-arm netbsd-arm64 openbsd-amd64 openbsd-arm openbsd-arm64 windows-amd64 windows-arm
DOCKER_PLATFORMS     ?= linux/amd64 linux/arm64
DOCKER_REPO          ?= mattolenik
DOCKER_TAG           ?= $(DOCKER_REPO)/cloudflare-ddns-client
DOCKER_TAG_LATEST    := $(DOCKER_TAG):latest
DOCKER_TAG_VERSIONED := $(DOCKER_TAG):$(RELEASE_VERSION)

export TEST_BINARY := $(DIST)/$(BIN_NAME)

default: all shasums readme test

build: $(DIST)/$(BIN_NAME)
$(DIST)/$(BIN_NAME): $(SOURCE)
	go build $(ARGS) $(FLAGS) -ldflags="$(LDFLAGS)" -o $@

all: $(addprefix $(DIST)/$(BIN_NAME)-,$(PLATFORMS))

docker-publish:
	docker buildx build --push --tag $(DOCKER_TAG_LATEST) --tag $(DOCKER_TAG_VERSIONED) --platform linux/amd64,linux/arm64 .

clean:
	rm -rf dist && mkdir dist

$(DIST)/$(BIN_NAME)-%: GOOS   = $(word 1,$(subst -, ,$*))
$(DIST)/$(BIN_NAME)-%: GOARCH = $(word 2,$(subst -, ,$*))
$(DIST)/$(BIN_NAME)-%: $(SOURCE)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(ARGS) $(FLAGS) -ldflags="$(LDFLAGS)" -o $@

shasums: $(DIST)/$(BIN_NAME)-shasums
$(DIST)/$(BIN_NAME)-shasums: all
	cd $(DIST) && shasum -a 256 $(BIN_NAME)-* > $(BIN_NAME)-shasums

install:
	go install -ldflags=$(LDFLAGS)

readme: README.md
README.md: README.tpl.md
	go run tools/readmegen/main.go README.tpl.md > README.md

test: build
	@mkdir -p $(DIST)
	$(GOTESTSUM) --format testname --junitfile $(DIST)/test-results-$(UNAME).xml

coverage:
	go test -v -covermode=count -coverprofile=$(DIST)/coverage.out $(MODULE)/...

# TODO: systemd init script
fpm: $(DIST)/cloudflare-ddns-linux-amd64
	@mkdir -p $(DIST)/fpm
	cp $(DIST)/cloudflare-ddns-linux-amd64 $(DIST)/fpm/cloudflare-ddns
	fpm -f -s dir -t deb -n cloudflare-ddns -v $(VERSION) --license Unlicense \
	  --maintainer matt@olenik.me \
	  --package $(DIST)/fpm/cloudflare-ddns.deb \
	  --description "A robust, automatic dynamic DNS client for CloudFlare" \
	  --config-files /etc/cloudflare-ddns.toml.example cloudflare-ddns.toml.example=/etc/ \
	  $(DIST)/fpm/cloudflare-ddns=/usr/bin/

.PHONY: clean install test