GOLDFLAGS=-s -w -extldflags '-zrelro -znow'

.PHONY: all
all: build

.PHONY: dist/arm64/mauzr
dist/arm64/mauzr:
	GOARCH=arm64 go build -trimpath -ldflags "$(GOLDFLAGS)" -o $@ ./cmd/mauzr

.PHONY: dist/amd64/mauzr
dist/amd64/mauzr:
	GOARCH=amd64 go build -trimpath -ldflags "$(GOLDFLAGS)" -o $@ ./cmd/mauzr

.PHONY: build
build: dist/arm64/mauzr dist/amd64/mauzr

.PHONY: benchmark
benchmark:
	go test -bench=. ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: vet
vet:
	golangci-lint run ./...

.PHONY: download
download:
	go mod download

.PHONY: build-image
build-image:
	docker build -t docker.pkg.github.com/eqrx/mauzr/mauzr:latest -f build/Dockerfile .

.PHONY: push-image
push-image:
	docker push docker.pkg.github.com/eqrx/mauzr/mauzr:latest

.PHONY: fmt
fmt:
	gofmt -s -w .

