.PHONY: build run

default: all

all: build run

build:
	@go build .

run:
	@./je
