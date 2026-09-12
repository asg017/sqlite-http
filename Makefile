COMMIT=$(shell git rev-parse HEAD)
VERSION=$(shell cat VERSION)
DATE=$(shell date +'%FT%TZ%z')

VENDOR_SQLITE=$(shell pwd)/sqlite
GO_BUILD_LDFLAGS=-ldflags '-X main.Version=v$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)'
GO_BUILD_CGO_CFLAGS=CGO_ENABLED=1 CGO_CFLAGS="-DUSE_LIBSQLITE3" CPATH="$(VENDOR_SQLITE)"

ifeq ($(shell uname -s),Darwin)
CONFIG_DARWIN=y
else ifeq ($(OS),Windows_NT)
CONFIG_WINDOWS=y
else
CONFIG_LINUX=y
endif

ifdef CONFIG_DARWIN
LOADABLE_EXTENSION=dylib
endif
ifdef CONFIG_LINUX
LOADABLE_EXTENSION=so
endif
ifdef CONFIG_WINDOWS
LOADABLE_EXTENSION=dll
endif

ifdef python
PYTHON=$(python)
else
PYTHON=python3
endif

prefix=dist

TARGET_LOADABLE=$(prefix)/http0.$(LOADABLE_EXTENSION)
TARGET_STATIC=$(prefix)/http0.a

GO_FILES=./cookies.go ./settings.go ./do.go ./shared.go ./meta.go ./headers.go

loadable: $(TARGET_LOADABLE)
static: $(TARGET_STATIC)
all: loadable

$(prefix):
	mkdir -p $(prefix)

$(TARGET_LOADABLE): $(GO_FILES) | $(prefix)
	$(GO_BUILD_CGO_CFLAGS) go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	$(GO_BUILD_LDFLAGS) \
	.

$(TARGET_STATIC): $(GO_FILES) | $(prefix)
	$(GO_BUILD_CGO_CFLAGS) go build \
	-buildmode=c-archive -o $@ \
	$(GO_BUILD_LDFLAGS) \
	.

format:
	gofmt -s -w .

clean:
	rm -rf $(prefix)

# Local httpbin for the test suite, the same server CI uses. No docker needed.
httpbin:
	go run github.com/mccutchen/go-httpbin/v2/cmd/go-httpbin@v2.25.0 -port 8080

test-loadable:
	$(PYTHON) tests/test-loadable.py

test-watch:
	watchexec --clear -w tests/test-loadable.py make test-loadable

test: test-loadable

# Build every distributable (npm, pip, gem, ...) from the extensions in
# dist/<target>/. CI populates those directories from the build matrix.
sqlite-dist: sqlite-dist.toml
	sqlite-dist build --set-version $(VERSION)

publish-release:
	./scripts/publish_release.sh

.PHONY: all loadable static format clean httpbin \
	test test-loadable test-watch sqlite-dist publish-release
