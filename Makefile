-include user/.env
export

export USER_PROJECT_ROOT=$(shell pwd)/user
export PROJECT_ROOT=$(shell pwd)

.PHONY: proto-gen run test lint up down migrate-up migrate-down

env-cleanup:
	@read -p "Clean up volume? [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down postgres port-forwarder && \
		rm -rf ${USER_PROJECT_ROOT}/out/postgres_data && \
		echo "Volume has been cleaned up"; \
	else \
		echo "Clean up cancelled"; \
	fi

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

	@docker compose run --rm migrate \
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

user-run:
	cd ${PROJECT_ROOT} && \
	go work sync && \
	cd ${USER_PROJECT_ROOT} && \
	go fmt ./... && \
	go run ./cmd