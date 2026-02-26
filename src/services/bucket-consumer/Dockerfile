# Stage 1 — Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bucket-consumer ./cmd/consumer

# Stage 2 — Runtime
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/bucket-consumer .

USER appuser

CMD ["./bucket-consumer"]
