-include user/.env
-include url/.env
-include gateway/.env
-include analytics/.env
export

export USER_PROJECT_ROOT=$(shell pwd)/user
export URL_PROJECT_ROOT=$(shell pwd)/url
export GATEWAY_PROJECT_ROOT=$(shell pwd)/gateway
export ANALYTICS_PROJECT_ROOT=$(shell pwd)/analytics
export PROJECT_ROOT=$(shell pwd)

.PHONY: proto-gen run test lint up down \
        migrate-up migrate-down \
        clickhouse-migrate-up clickhouse-migrate-down clickhouse-migrate-action clickhouse-migrate-create \
        url-run gateway-run analytics-run run-all

env-cleanup:
	@read -p "Clean up volume? [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down postgres port-forwarder && \
		rm -rf ${USER_PROJECT_ROOT}/out/postgres_data && \
		rm -rf ${URL_PROJECT_ROOT}/out/mongo_data && \
		rm -rf ${ANALYTICS_PROJECT_ROOT}/out/clickhouse_data && \
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

test-int:
	go test -count=1 -timeout=15m -tags=integration ./gateway/integration/...

lint:
	golangci-lint run ./user/...

up:
	docker compose up -d --build
	@$(MAKE) migrate-up
	@$(MAKE) clickhouse-migrate-up

down:
	docker compose down --remove-orphans

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "action is not specified. Example: make migrate-action action=up"; \
		exit 1; \
	fi;
	@docker compose run --rm migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

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

clickhouse-migrate-up:
	@make clickhouse-migrate-action action=up

clickhouse-migrate-down:
	@make clickhouse-migrate-action action=down

clickhouse-migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "action is not specified. Example: make clickhouse-migrate-action action=up"; \
		exit 1; \
	fi
	@docker compose run --rm clickhouse-migrate \
		-path=/migrations \
		-database "clickhouse://clickhouse:9000?username=${CLICKHOUSE_USERNAME}&password=${CLICKHOUSE_PASSWORD}&database=${CLICKHOUSE_DATABASE}&x-multi-statement=true" \
		"$(action)"

clickhouse-migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "seq parameter is not specified. Example: make clickhouse-migrate-create seq=create-events"; \
		exit 1; \
	fi
	@docker compose run --rm clickhouse-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

swagger-gen:
	@docker compose run --rm swagger \
		init \
		-g /cmd/main.go \
		-o docs \
		--parseInternal \
		--parseDependency

user-run:
	@cd ${USER_PROJECT_ROOT} && \
	go fmt ./... && \
	go run ./cmd

url-run:
	@cd ${URL_PROJECT_ROOT} && \
	go fmt ./... && \
	go run ./cmd

gateway-run:
	@cd ${GATEWAY_PROJECT_ROOT} && \
	go fmt ./... && \
	go run ./cmd

analytics-run:
	@cd ${ANALYTICS_PROJECT_ROOT} && \
	go fmt ./... && \
	go run ./cmd

run-all: 
	@$(MAKE) migrate-up
	@$(MAKE) clickhouse-migrate-up
	@$(MAKE) user-run & \
	$(MAKE) url-run & \
	$(MAKE) analytics-run & \
	$(MAKE) gateway-run
