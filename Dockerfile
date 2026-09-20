# --- Build stage (tetap sama) ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# --- Run stage ---
FROM alpine:latest

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migration ./migration

RUN chown -R appuser /app
USER appuser

EXPOSE 8080

CMD ["./main"]