FROM golang:1.27.3-alpine AS builder

WORKDIR /src

COPY go.work* ./
COPY shared/go.mod shared/go.sum ./shared/
COPY user/go.mod user/go.sum ./user/
COPY url/go.mod url/go.sum ./url/
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY analytics/go.mod analytics/go.sum ./analytics/

COPY shared ./shared
COPY user ./user
COPY url ./url
COPY gateway ./gateway
COPY analytics ./analytics

RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/url ./url/cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/url /app/url

EXPOSE 50052 9091

ENTRYPOINT ["/app/url"]
