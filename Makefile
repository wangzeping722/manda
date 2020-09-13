.PHONY: proto all

proto:
	protoc ./proto/*.proto --go_out=:./

all: proto
	go build -o bin/manda cmd/*
