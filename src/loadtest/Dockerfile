# ─── Stage 1: Compilar k6 com xk6-kafka ────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

# Instala o xk6
RUN go install go.k6.io/xk6/cmd/xk6@latest

# Compila o binário k6 com a extensão Kafka
RUN xk6 build \
    --with github.com/mostafa/xk6-kafka@latest \
    --output /k6

# ─── Stage 2: Imagem final enxuta ───────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S k6 && adduser -S k6 -G k6

COPY --from=builder /k6 /usr/local/bin/k6
COPY loadtest.js     /scripts/loadtest.js

RUN chown -R k6:k6 /scripts

USER k6
WORKDIR /scripts

ENTRYPOINT ["k6"]
CMD ["run", "/scripts/loadtest.js"]