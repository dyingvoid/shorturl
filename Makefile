-include user/.env
export

export USER_PROJECT_ROOT=$(shell pwd)/user

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

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "seq parameter is not specified. Example: make migrate-create seq=init"; \
		exit 1; \
	fi;

	@docker compose run --rm todo-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "action is not specified. Example: make migrate-action action=up"; \
		exit 1; \
	fi;
	@docker compose run --rm todo-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"