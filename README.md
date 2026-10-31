# ShortURL

URL shortener built as a set of Go services behind an HTTP gateway:

- **gateway** — HTTP API (port `8082`), validates requests and calls the gRPC services below.
- **user** — registration, login, sessions (gRPC `50051`).
- **url** — short link creation, redirect, listing (gRPC `50052`).
- **analytics** — click statistics (gRPC `50053`).
- **Infrastructure** — PostgreSQL, MongoDB, ClickHouse, Redis, Kafka.

Requests go through the gateway; it is the only service exposed to API clients.
The redirect route lives outside `/api/v1`, so short links look like
`http://localhost:8082/<code>`.

## Launch

Copy the environment files first if they are missing:

```bash
cp user/.env.example user/.env
cp url/.env.example url/.env
cp analytics/.env.example analytics/.env
cp gateway/.env.example gateway/.env
```

Fill in the empty values (at minimum ports and service hosts) and start everything:

```bash
make up
```

`make up` builds and starts all containers with Docker Compose, then applies
PostgreSQL and ClickHouse migrations. The gateway is available at
`http://localhost:8082`.

Other useful targets:

```bash
make down        # stop containers
make run-all     # run all services locally with go run (infra must be up)
make gateway-run # run only the gateway locally
```

Follow the logs with `docker compose logs -f gateway` (or `user`, `url`, `analytics`).

For local development, run the infrastructure in Docker and then `make run-all`:

```bash
docker compose up -d postgres mongo clickhouse redis kafka jaeger
make run-all
```

Note: `make run-all` runs `go run` directly, so either stop the `user`, `url`,
`analytics` and `gateway` containers or point the env files at the local ports
and non-conflicting host ports.

## Swagger

Swagger UI is served by the gateway (it is embedded in the gateway image):

- UI: <http://localhost:8082/swagger/index.html>
- OpenAPI document: <http://localhost:8082/swagger/doc.json>

Swagger annotations are generated from the gateway handlers into
`gateway/docs`. To regenerate after changing annotations:

```bash
make swagger-gen
```

The generated spec is checked in, so this is only needed when annotations change.

## API overview

All versioned routes are under `/api/v1`. Routes marked **auth** require the
access token from login in the header, using this exact format:

```text
Authorization: Bearer %access_token%
```

For example:

```bash
curl http://localhost:8082/api/v1/urls \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
```

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/register` | no | Create a user |
| POST | `/api/v1/auth/login` | no | Get access + refresh tokens |
| POST | `/api/v1/auth/refresh` | no | Exchange refresh token for a new access token |
| POST | `/api/v1/auth/logout` | yes | Invalidate the current session |
| POST | `/api/v1/urls` | yes | Create a short URL |
| GET | `/api/v1/urls` | yes | List the user's URLs (`limit`, `cursor`) |
| DELETE | `/api/v1/urls/{code}` | yes | Delete a URL |
| GET | `/api/v1/analytics/{code}` | yes | Click analytics (`days`, default `30`) |
| GET | `/{code}` | no | Redirect to the original URL (`302`) |

## Request flow

1. **Register**

   ```bash
   curl -X POST http://localhost:8082/api/v1/auth/register \
     -H 'Content-Type: application/json' \
     -d '{"email":"user@example.com","password":"Secret1"}'
   ```

   Response: `201` with `{"user_id":"S..."}`.

2. **Log in** and keep the tokens:

   ```bash
   curl -X POST http://localhost:8082/api/v1/auth/login \
     -H 'Content-Type: application/json' \
     -d '{"email":"user@example.com","password":"Secret1"}'
   ```

   Response: `200` with `access_token` and `refresh_token`. Use the access
   token in subsequent requests as `Authorization: Bearer %access_token%`.

3. **Create a short URL** using the access token:

   ```bash
   curl -X POST http://localhost:8082/api/v1/urls \
     -H 'Authorization: Bearer %access_token%' \
     -H 'Content-Type: application/json' \
     -d '{"original_url":"https://example.com/some/long/path"}'
   ```

   Response: `201` with `short_code` and `short_url`
   (for example `http://localhost:8082/abc123`).

4. **Follow the short link** (no auth, no `/api/v1` prefix):

   ```bash
   curl -i http://localhost:8082/abc123
   ```

   The gateway calls the URL service, which records the click by publishing an
   event to Kafka, then returns `302` with a `Location` header.

5. **Analytics** — the analytics service consumes click events from Kafka and
   stores them in ClickHouse; query them through the gateway:

   ```bash
   curl http://localhost:8082/api/v1/analytics/abc123?days=30 \
     -H 'Authorization: Bearer %access_token%'
   ```

   Response: `short_code`, `total_clicks`, `clicks_by_day` and `top_referers`.

6. **Manage links** — list with `GET /api/v1/urls`, delete with
   `DELETE /api/v1/urls/abc123`. Both require the bearer token
   (`Authorization: Bearer %access_token%`).

7. **Refresh / log out** — use `POST /api/v1/auth/refresh` with
   `{"refresh_token":"..."}` to obtain a new access token, and
   `POST /api/v1/auth/logout` with the bearer token to end the session.

## Monitoring

The Docker Compose setup also starts Prometheus (`http://localhost:9090`),
Grafana (`http://localhost:3000`, default `admin`/`admin`),
Jaeger (`http://localhost:16686`) and Loki. Each Go service exposes
`/metrics` and `/debug/pprof` on its metrics port (`9090` inside the
containers, mapped to `9100`+ on the host).
