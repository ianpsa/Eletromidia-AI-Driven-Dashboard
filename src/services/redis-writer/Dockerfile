# ---------- build stage ----------
FROM golang:1.25.6-alpine AS builder

WORKDIR /.

# Dependências do Go (certs e git pra baixar módulos)
RUN apk add --no-cache ca-certificates git

# Copia apenas mod/sum primeiro (melhora cache do build)
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Build estático (boa prática pra rodar em imagem final minimalista)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o consumer ./


# ---------- runtime stage ----------
FROM alpine:3.20

WORKDIR /.

# Certificados pra TLS (se seu Kafka usar SASL_SSL/SSL)
RUN apk add --no-cache ca-certificates

# Usuário não-root (boa prática)
RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /consumer /consumer

ENTRYPOINT ["/consumer"]
