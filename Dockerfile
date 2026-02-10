# Build Stage
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/api ./cmd/api
RUN go build -o /bin/mqtt-subscriber ./cmd/mqtt-subscriber
RUN go build -o /bin/worker ./cmd/worker

# Run Stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/api /bin/api
COPY --from=builder /bin/mqtt-subscriber /bin/mqtt-subscriber
COPY --from=builder /bin/worker /bin/worker

WORKDIR /app