COMMIT=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --tags --exact-match --always)
VERSION=v0.0.0
DATE=$(shell date +'%FT%TZ%z')

GO_BUILD_LDFLAGS=-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)' 
GO_BUILD_NO_NET_LDFLAGS=-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE) -X main.OmitNet=1'
GO_BUILD_CGO_CFLAGS=CGO_CFLAGS=-DSQLITE3_INIT_FN=sqlite3_http_init
GO_BUILD_NO_NET_CGO_CFLAGS=CGO_CFLAGS=-DSQLITE3_INIT_FN=sqlite3_httpnonet_init

CGO_CFLAGS="-DSQLITE3_INIT_FN=sqlite3_httpnonet_init"

ifeq ($(OS),Windows_NT)
CONFIG_WINDOWS=y
endif

ifeq ($(shell uname -s),Darwin)
CONFIG_DARWIN=y
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


TARGET_LOADABLE=dist/http0.$(LOADABLE_EXTENSION)
TARGET_LOADABLE_NO_NET=dist/http0-no-net.$(LOADABLE_EXTENSION)
TARGET_OBJ=dist/http0.o
TARGET_SQLITE3=dist/sqlite3

loadable: $(TARGET_LOADABLE) $(TARGET_LOADABLE_NO_NET)
sqlite3: $(TARGET_SQLITE3)
package: $(TARGET_PACKAGE)
all: loadable sqlite3 package

$(TARGET_LOADABLE):  $(shell find . -type f -name '*.go')
	$(GO_BUILD_CGO_CFLAGS) go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	$(GO_BUILD_LDFLAGS) \
	.

$(TARGET_LOADABLE_NO_NET):  $(shell find . -type f -name '*.go')
	$(GO_BUILD_NO_NET_CGO_CFLAGS) go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	$(GO_BUILD_NO_NET_LDFLAGS) \
	.

$(TARGET_OBJ):  $(shell find . -type f -name '*.go')
	$(GO_BUILD_CGO_CFLAGS) CGO_ENABLED=1 go build -buildmode=c-archive \
	$(GO_BUILD_LDFLAGS) \
	-o $@ .

# framework stuff is needed bc https://github.com/golang/go/issues/42459#issuecomment-896089738
$(TARGET_SQLITE3): $(TARGET_OBJ) dist/sqlite3-extra.c sqlite/shell.c
	gcc \
	-framework CoreFoundation -framework Security \
	dist/sqlite3-extra.c sqlite/shell.c $(TARGET_OBJ) \
	-L. -Isqlite \
	-DSQLITE_EXTRA_INIT=core_init -DSQLITE3_INIT_FN=sqlite3_http_init \
	-o $@

$(TARGET_PACKAGE): $(TARGET_LOADABLE) $(TARGET_OBJ) sqlite/sqlite-http.h $(TARGET_SQLITE3)
	zip --junk-paths $@ $(TARGET_LOADABLE) $(TARGET_OBJ) sqlite/sqlite-http.h $(TARGET_SQLITE3)

dist/sqlite3-extra.c: sqlite/sqlite3.c sqlite/core_init.c
	cat sqlite/sqlite3.c sqlite/core_init.c > $@

format:
	gofmt -s -w .

httpbin: 
	docker run -p 8080:80 kennethreitz/httpbin

clean:
	rm dist/*

test: $(TARGET_LOADABLE) $(TARGET_LOADABLE_NO_NET)
	python3 test.py

test-watch:
	watchexec --clear -w test.py make test


.PHONY: all format clean \
	test test-watch httpbin \
	loadable sqlite3