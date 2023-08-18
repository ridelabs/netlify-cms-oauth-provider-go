#TARGET_BIN = netlify-cms-oauth-provider
#TARGET_ARCH = amd64
#SOURCE_MAIN = main.go
#LDFLAGS = -s -w
#
#all: build
#
#build: build-darwin build-linux build-windows
#
#build-darwin:
#	CGO_ENABLED=0 GOOS=darwin GOARCH=$(TARGET_ARCH) go build -ldflags "$(LDFLAGS)" -o bin/$(TARGET_BIN)_darwin-amd64 $(SOURCE_MAIN)
#
#build-linux:
#	CGO_ENABLED=0 GOOS=linux GOARCH=$(TARGET_ARCH) go build -ldflags "$(LDFLAGS)" -o bin/$(TARGET_BIN)_linux-amd64 $(SOURCE_MAIN)
#
#build-windows:
#	CGO_ENABLED=0 GOOS=windows GOARCH=$(TARGET_ARCH) go build -ldflags "$(LDFLAGS)" -o bin/$(TARGET_BIN)_windows-amd64.exe $(SOURCE_MAIN)
#
#start:
#	go run $(SOURCE_MAIN)

.PHONY: all deps clean docs test fmt lint install

#TAGS =
TAGS = osusergo,netgo

INSTALL_DIR        = $(GOPATH)/bin
WEBSITE_DIR        = ./website
DEST_DIR           = ./target
PATHINSTBIN        = $(DEST_DIR)/bin


#VERSION   := $(shell git describe --tags || echo "v0.0.0")
VERSION   := v0.0.0
VER_CUT   := $(shell echo $(VERSION) | cut -c2-)
VER_MAJOR := $(shell echo $(VER_CUT) | cut -f1 -d.)
VER_MINOR := $(shell echo $(VER_CUT) | cut -f2 -d.)
VER_PATCH := $(shell echo $(VER_CUT) | cut -f3 -d.)
VER_RC    := $(shell echo $(VER_PATCH) | cut -f2 -d-)
DATE      := $(shell date +"%Y-%m-%dT%H:%M:%SZ")


LD_FLAGS   = -w -s
#
# when building for destination in DigitalOcean we have to pass --buildmode=pie
#GO_FLAGS   = -buildmode=pie
GO_FLAGS =

DOCS_FLAGS =

APPS = netlify-cms-oauth-provider


all: $(APPS)

deps:
	@go mod tidy

clean:
	rm -rf $(PATHINSTBIN)

SOURCE_FILES = $(shell find internal public cmd -type f)
TEMPLATE_FILES = $(shell find template -path template/test -prune -o -type f -name "*.yaml")


$(PATHINSTBIN)/%: $(SOURCE_FILES) $(TEMPLATE_FILES)
	go build $(GO_FLAGS) -tags "$(TAGS)" -ldflags "$(LD_FLAGS) $(VER_FLAGS)" -o $@ ./cmd/$*
	#go build $(GO_FLAGS) -tags "$(TAGS)" -ldflags "$(LD_FLAGS) $(VER_FLAGS)" -o $@ ./cmd/$*
	#go build -o $@ ./cmd/$*

$(APPS): %: $(PATHINSTBIN)/%

