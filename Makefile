COMMIT=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --tags --exact-match --always)
DATE=$(shell date +'%FT%TZ%z')

dist/http.so:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)' \
	shared.go