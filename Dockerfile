# --- Stage 1: Builder ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main .

# --- Stage 2: Final Image ---
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
