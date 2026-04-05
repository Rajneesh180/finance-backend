.PHONY: build run test lint migrate-up migrate-down docker-up docker-down

APP_NAME := finance-backend
DB_URL ?= postgres://postgres:postgres@localhost:5432/finance?sslmode=disable

build:
	go build -o bin/$(APP_NAME) ./cmd/server/

run: build
	JWT_SECRET=dev-secret ./bin/$(APP_NAME)

test:
	go test -v -race -count=1 ./...

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path internal/database/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path internal/database/migrations -database "$(DB_URL)" down 1

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
