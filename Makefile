.PHONY: build run test lint clean

clean:
	@rm -rf bin

build:
	@go build -o bin/api ./cmd/api

run: clean build
	@./bin/api