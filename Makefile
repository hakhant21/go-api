.PHONY: run build test test-cover tidy swagger migrate-diff migrate-apply \
        migrate-status docker-up docker-down docker-logs docker-reset

run:
	go run ./cmd/api

build:
	go build -ldflags="-s -w" -o bin/api ./cmd/api

test:
	go test ./tests/... -race -count=1

test-cover:
	go test ./tests/... -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

test-all:
	go test ./... -race -count=1

tidy:
	go mod tidy

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs --parseDependency --parseInternal

migrate-diff:
	atlas migrate diff $(name) --env local

migrate-apply:
	atlas migrate apply --env local --url "postgres://postgres:postgres@localhost:5432/starter?sslmode=disable"

migrate-status:
	atlas migrate status --env local --url "postgres://postgres:postgres@localhost:5432/starter?sslmode=disable"

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api

docker-reset:
	docker compose down -v
	docker compose up -d --build
