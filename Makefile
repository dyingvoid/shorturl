-include user/.env
export

.PHONY: proto-gen run test lint up down migrate-up migrate-down

proto-gen:
	cd shared/proto && buf generate

run:
	go run ./user/cmd

test:
	go test -race -v ./user/...

lint:
	golangci-lint run ./user/...

up:
	docker compose up -d

down:
	docker compose down

migrate-up:
	migrate -path user/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path user/migrations -database "$(DATABASE_URL)" down 1
