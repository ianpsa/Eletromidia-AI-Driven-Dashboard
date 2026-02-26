# Bucket Consumer

Consumer Kafka em Go que persiste eventos do tópico `geodata` no Google Cloud Storage como backup em NDJSON.

## Estrutura

```
bucket-consumer/
├── cmd/
│   ├── consumer/main.go          # Entrypoint do serviço
│   └── produce-test/main.go      # Ferramenta para enviar mensagem de teste
├── internal/
│   ├── config/config.go          # Configuração via variáveis de ambiente
│   ├── consumer/consumer.go      # Loop de consumo Kafka
│   └── storage/writer.go         # Buffer + flush para GCS
├── go.mod
├── go.sum
├── Dockerfile
├── .dockerignore
├── .env.example
└── docker-compose.yml
```

- `cmd/` — Binários do projeto. Cada subdiretório produz um executável.
- `internal/` — Código privado. Não pode ser importado por outros módulos Go.

## Organização no bucket

```
gs://kafka-backup-eletromidia/
└── kafka-backup/
    └── topics/
        └── geodata/
            └── year=2026/
                └── month=02/
                    └── day=26/
                        └── {partition}-{startOffset}-{endOffset}.json
```

## Pré-requisitos

- Go 1.25+
- Docker e Docker Compose
- Kafka rodando em `localhost:9092`
- Arquivo `cs-api-key.json` com credenciais de Service Account do GCS

## Como rodar

### 1. Configurar o `.env`

```bash
cp .env.example .env
```

Ajuste `GCS_CREDENTIALS` se o arquivo de credenciais estiver em outro caminho. O `.env` é usado tanto pelo `go run` quanto pelo `docker compose`.

> Para testes rápidos, use `FLUSH_SIZE=1` para forçar flush a cada mensagem.

### 2. Via Go

```bash
go run cmd/consumer/main.go
```

### 3. Via Docker

```bash
docker compose up --build
```

O `docker-compose.yml` faz três coisas:
- **Builda** a imagem a partir do `Dockerfile`
- **Carrega** as variáveis do `.env` via `env_file`
- **Monta** o `cs-api-key.json` como volume read-only no container

O compose usa `network_mode: host`, o que significa que o container **compartilha a rede da máquina**. Isso permite que o mesmo `.env` com `KAFKA_BROKERS=localhost:9092` funcione tanto via `go run` quanto via `docker compose`, sem precisar de configurações de rede separadas.

Para parar:

```bash
docker compose down
```

Para produção (Kubernetes), basta alterar `KAFKA_BROKERS` no `.env`:

```
KAFKA_BROKERS=my-cluster-kafka-bootstrap.kafka.svc:9092
```

## Enviando mensagem de teste

Em **outro terminal**, no mesmo diretório:

```bash
go run cmd/produce-test/main.go
```

Saída esperada:

```
mensagem enviada com sucesso para o topico geodata
```

No terminal do consumer:

```
bufferizado | particao=0 offset=0 pendentes=1
flush concluido: 1 registros → gs://kafka-backup-eletromidia/kafka-backup/topics/geodata/year=2026/month=02/day=26/0-0-0.json
```

## Variáveis de ambiente

| Variável | Descrição | Default |
|---|---|---|
| `KAFKA_BROKERS` | Endereço(s) do Kafka | `my-cluster-kafka-bootstrap.kafka.svc:9092` |
| `KAFKA_TOPIC` | Tópico para consumir | `geodata` |
| `KAFKA_GROUP_ID` | Consumer group ID | `go-consumer-group` |
| `KAFKA_MIN_BYTES` | Mínimo de bytes por fetch | `1000` |
| `KAFKA_MAX_BYTES` | Máximo de bytes por fetch | `10000000` |
| `KAFKA_READ_TIMEOUT` | Timeout por leitura de mensagem | `5s` |
| `KAFKA_MAX_WAIT` | Tempo máximo de espera por batch do broker | `500ms` |
| `PROCESS_DELAY` | Delay entre processamento de mensagens | `0` |
| `GCS_BUCKET` | Nome do bucket GCS | `kafka-backup-eletromidia` |
| `GCS_CREDENTIALS` | Caminho para o JSON de credenciais | (usa default credentials se vazio) |
| `GCS_BASE_PATH` | Prefixo do path no bucket | `kafka-backup` |
| `FLUSH_INTERVAL` | Intervalo máximo entre flushes | `30s` |
| `FLUSH_SIZE` | Quantidade de mensagens para disparar flush | `500` |
