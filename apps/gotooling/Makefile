


DIR=$(shell realpath $(shell pwd))
BIN_DIR ?= $(shell realpath "${DIR}/../../bin")


build:
	go build -o $(BIN_DIR) -ldflags="-s -w" -tags netgo ./...

