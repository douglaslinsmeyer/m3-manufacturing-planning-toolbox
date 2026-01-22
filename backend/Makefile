.PHONY: help build run test clean migrate-up migrate-down migrate-create docker-build docker-run

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make migrate-create - Create a new migration (NAME=migration_name)"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run with Docker Compose"

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean

migrate-up:
	go run cmd/server/main.go migrate up

migrate-down:
	go run cmd/server/main.go migrate down

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(NAME)

docker-build:
	docker-compose build

docker-run:
	docker-compose up -d
