# ---- Stage 1: build ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ticket-system ./cmd/server

# ---- Stage 2: runtime ----
FROM alpine:3.20


RUN adduser -D -u 10001 appuser
USER appuser

COPY --from=builder /ticket-system /ticket-system

EXPOSE 8080

ENTRYPOINT ["/ticket-system"]
