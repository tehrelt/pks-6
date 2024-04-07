server:
	go build -o ./bin -v ./cmd/server

client:
	go build -o ./bin -v ./cmd/client

build:
	make server
	make client

.DEFAULT_GOAL := build