# ETAP 1: Budowanie
FROM golang:1.26.2-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/bin/server ./cmd/main

# ETAP 2: Obraz docelowy
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bin/server .

RUN useradd -u 1001 appuser
USER appuser

EXPOSE 50051
EXPOSE 8081

CMD ["./server"]