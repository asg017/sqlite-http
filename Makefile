COMMIT=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --tags --exact-match --always)
VERSION=v0.0.0
DATE=$(shell date +'%FT%TZ%z')


dist/http0-macos.dylib:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)' \
	shared.go

dist/http0-macos-no-net.dylib:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE) -X main.OmitNet=1' \
	shared.go


dist/http0-linux.so:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)' \
	shared.go

dist/http0-linux-no-net.so:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE) -X main.OmitNet=1' \
	shared.go

dist/http0-windows.dll:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)' \
	shared.go

dist/http0-windows-no-net.dll:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE) -X main.OmitNet=1' \
	shared.go

macos: dist/http0-macos.dylib dist/http0-macos-no-net.dylib

linux: dist/http0-linux.so dist/http0-linux-no-net.so

windows: dist/http0-windows.dll dist/http0-windows-no-net.dll

format:
	gofmt -s -w .

httpbin: 
	docker run -p 8080:80 kennethreitz/httpbin

test:
	python3 test.py

test-watch:
	watchexec --clear -w test.py make test

.PHONY: httpbin all test test-watch format macos linux windows