export GITTAG=$(shell git describe --tags --always)
export GITCOMMIT=$(shell git log -1 --pretty=format:"%H")
export GOLDFLAGS=-s -w -extldflags '-zrelro -znow' -X go.eqrx.net/mauzr.version=$(GITTAG) -X go.eqrx.net/mauzr.commit=$(GITCOMMIT)
export GOFLAGS=-trimpath
export CGO_ENABLED=0

.PHONY: all
all: build

.PHONY: build
build:
	go build -ldflags "$(GOLDFLAGS)" ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem ./...

.PHONY: test
test:
	go test -timeout 5s -race ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: download
download:
	go mod download

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: update
update:
	go get -t -u=patch ./...
