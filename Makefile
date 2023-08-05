COMMIT=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --tags --exact-match --always)
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

# framework stuff is needed bc https://github.com/golang/go/issues/42459#issuecomment-896089738
ifdef CONFIG_DARWIN
LOADABLE_EXTENSION=dylib
SQLITE3_CFLAGS=-framework CoreFoundation -framework Security
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

ifdef IS_MACOS_ARM
RENAME_WHEELS_ARGS=--is-macos-arm
else
RENAME_WHEELS_ARGS=
endif

prefix=dist

TARGET_LOADABLE=$(prefix)/http0.$(LOADABLE_EXTENSION)
TARGET_WHEELS=$(prefix)/wheels
TARGET_OBJ=$(prefix)/http0.o
TARGET_SQLITE3=$(prefix)/sqlite3

INTERMEDIATE_PYPACKAGE_EXTENSION=python/sqlite_http/sqlite_http/http0.$(LOADABLE_EXTENSION)

loadable: $(TARGET_LOADABLE)
all: loadable

GO_FILES= ./cookies.go ./settings.go ./do.go ./shared.go ./meta.go ./headers.go

$(prefix):
	mkdir -p $(prefix)

$(TARGET_WHEELS): $(prefix)
	mkdir -p $(TARGET_WHEELS)

$(TARGET_LOADABLE):  $(GO_FILES)
	$(GO_BUILD_CGO_CFLAGS) go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	$(GO_BUILD_LDFLAGS) \
	.

python: $(TARGET_WHEELS) $(TARGET_LOADABLE) $(TARGET_WHEELS) scripts/rename-wheels.py $(shell find python/sqlite_http -type f -name '*.py')
	cp $(TARGET_LOADABLE) $(INTERMEDIATE_PYPACKAGE_EXTENSION)
	rm $(TARGET_WHEELS)/sqlite_http* || true
	pip3 wheel python/sqlite_http/ -w $(TARGET_WHEELS)
	python3 scripts/rename-wheels.py $(TARGET_WHEELS) $(RENAME_WHEELS_ARGS)
	echo "✅ generated python wheel"

python-versions: python/version.py.tmpl
	VERSION=$(VERSION) envsubst < python/version.py.tmpl > python/sqlite_http/sqlite_http/version.py
	echo "✅ generated python/sqlite_http/sqlite_http/version.py"

	VERSION=$(VERSION) envsubst < python/version.py.tmpl > python/datasette_sqlite_http/datasette_sqlite_http/version.py
	echo "✅ generated python/datasette_sqlite_http/datasette_sqlite_http/version.py"

datasette: $(TARGET_WHEELS) $(shell find python/datasette_sqlite_http -type f -name '*.py')
	rm $(TARGET_WHEELS)/datasette* || true
	pip3 wheel python/datasette_sqlite_http/ --no-deps -w $(TARGET_WHEELS)

npm: VERSION npm/platform-package.README.md.tmpl npm/platform-package.package.json.tmpl npm/sqlite-http/package.json.tmpl scripts/npm_generate_platform_packages.sh
	scripts/npm_generate_platform_packages.sh

deno: VERSION deno/deno.json.tmpl
	scripts/deno_generate_package.sh

bindings/ruby/lib/version.rb: bindings/ruby/lib/version.rb.tmpl VERSION
	VERSION=$(VERSION) envsubst < $< > $@

ruby: bindings/ruby/lib/version.rb

version:
	make python
	make python-versions
	make npm
	make deno
	make ruby

$(TARGET_OBJ):  $(GO_FILES)
	$(GO_BUILD_CGO_CFLAGS) CGO_ENABLED=1 go build -buildmode=c-archive \
	$(GO_BUILD_LDFLAGS) \
	-o $@ .

dist/sqlite3-extra.c: sqlite/sqlite3.c sqlite/core_init.c
	cat sqlite/sqlite3.c sqlite/core_init.c > $@

format:
	gofmt -s -w .

httpbin:
	docker run -p 8080:80 kennethreitz/httpbin

clean:
	rm dist/*

test-loadable:
	$(PYTHON) tests/test-loadable.py

test-python:
	$(PYTHON) tests/test-python.py

test-npm:
	node npm/sqlite-http/test.js

test-deno:
	deno task --config deno/deno.json test

test-watch:
	watchexec --clear -w tests/test-loadable.py make test-loadable

test:
	make test-loadable
	make test-python
	make test-npm
	make test-deno

publish-release:
	./scripts/publish_release.sh

.PHONY: all format clean publish-release \
	python python-versions datasette npm deno ruby version \
	test test-loadable test-watch httpbin \
	loadable
