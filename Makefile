-include .env
export

.PHONY: build run test fmt vet clean

build:
	go build -o bot.bin ./cmd/bot

run: build
	./bot.bin

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f bot.bin
