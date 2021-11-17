COMMIT=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --tags --exact-match --always)
VERSION="v0.0.0"
DATE=$(shell date +'%FT%TZ%z')

all: dist/http0.so

dist/http0.so:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)' \
	shared.go

dist/http0-no-do.so:  $(shell find . -type f -name '*.go')
	go build \
	-buildmode=c-shared -o $@ -tags="shared" \
	-ldflags '-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE) -X main.OmitDo=1' \
	shared.go

httpbin: 
	docker run -p 8080:80 kennethreitz/httpbin

test:
	./test.sh 

format:
	gofmt -s -w .

.PHONY: httpbin all test format