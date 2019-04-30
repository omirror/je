.PHONY: dev build deps generate install image release profile bench test clean

CGO_ENABLED=0
VERSION=$(shell git describe --abbrev=0 --tags)
COMMIT=$(shell git rev-parse --short HEAD)

all: dev

dev: build
	@./je -v
	@./job --version

build: clean generate
	@go build \
		-tags "netgo static_build" -installsuffix netgo \
		-ldflags "-w -X $(shell go list).Version=$(VERSION) -X $(shell go list).Commit=$(COMMIT)" \
		./cmd/je/...
	@go build \
		-tags "netgo static_build" -installsuffix netgo \
		-ldflags "-w -X $(shell go list).Version=$(VERSION) -X $(shell go list).Commit=$(COMMIT)" \
		./cmd/job/...

deps:
	@[ -f $(shell go env GOPATH)/bin/rice ] || go get -u github.com/GeertJohan/go.rice/rice

generate: deps
	@go generate $(shell go list)/...

install: build
	@go install ./cmd/je/...
	@go install ./cmd/job/...

image:
	@docker build -t prologic/je .

release:
	@./tools/release.sh

profile: build
	@go test -cpuprofile cpu.prof -memprofile mem.prof -v -bench .

bench: build
	@go test -v -benchmem -bench=. .

test: build
	@go test -v \
		-cover -coverprofile=coverage.txt -covermode=atomic \
		-coverpkg=$(shell go list) \
		-race \
		.

clean:
	@git clean -f -d -X
