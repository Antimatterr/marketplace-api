.PHONY: build run test lint clean

clean:
	@rm -rf bin

build:
	@go build -o bin/api ./cmd/api

run: clean build
	@./bin/api

migrate-up:
	@go run ./cmd/migrate up

migrate-down:
	@go run ./cmd/migrate down