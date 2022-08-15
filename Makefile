COMMIT=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --tags --exact-match --always)
VERSION=$(shell cat VERSION)
DATE=$(shell date +'%FT%TZ%z')

GO_BUILD_LDFLAGS=-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)' 
GO_BUILD_NO_NET_LDFLAGS=-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE) -X main.OmitNet=1'
GO_BUILD_CGO_CFLAGS=CGO_CFLAGS=-DSQLITE3_INIT_FN=sqlite3_http_init
GO_BUILD_NO_NET_CGO_CFLAGS=CGO_CFLAGS=-DSQLITE3_INIT_FN=sqlite3_httpnonet_init

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


TARGET_LOADABLE=dist/http0.$(LOADABLE_EXTENSION)
TARGET_LOADABLE_NO_NET=dist/http0-no-net.$(LOADABLE_EXTENSION)
TARGET_OBJ=dist/http0.o
TARGET_SQLITE3=dist/sqlite3

loadable: $(TARGET_LOADABLE) $(TARGET_LOADABLE_NO_NET)
all: loadable 

GO_FILES= ./cookies.go ./settings.go ./shared.go ./meta.go ./headers.go

$(TARGET_LOADABLE):  $(GO_FILES)
	CGO_CFLAGS="-DUSE_LIBSQLITE3" CPATH=/Users/alex/projects/sqlite-http go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	$(GO_BUILD_LDFLAGS) \
	.

$(TARGET_LOADABLE_NO_NET):  $(GO_FILES)
	$(GO_BUILD_NO_NET_CGO_CFLAGS) go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	$(GO_BUILD_NO_NET_LDFLAGS) \
	.

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

test-loadable: $(TARGET_LOADABLE) $(TARGET_LOADABLE_NO_NET)
	python3 tests/test-loadable.py

test-watch:
	watchexec --clear -w tests/test-loadable.py make test-loadable

test:
	make test-loadable

.PHONY: all format clean \
	test test-loadable test-watch httpbin \
	loadable