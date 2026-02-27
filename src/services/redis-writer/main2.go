// Arquivo inicial somente para teste

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers       []string
	Topic         string
	GroupID       string
	MinBytes      int
	MaxBytes      int
	MaxWait       time.Duration
	ReadTimeout   time.Duration
	ProcessDelay  time.Duration // só pra simular processamento
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func loadConfig() Config {
	brokers := getEnv("KAFKA_BROKERS", "my-cluster-kafka-bootstrap.kafka.svc:9092")
	return Config{
		Brokers:      splitComma(brokers),
		Topic:        getEnv("KAFKA_TOPIC", "outro-topico"),
		GroupID:      getEnv("KAFKA_GROUP_ID", "go-consumer-group2"),
		MinBytes:     getEnvInt("KAFKA_MIN_BYTES", 1e3),   // 1KB
		MaxBytes:     getEnvInt("KAFKA_MAX_BYTES", 10e6),  // 10MB
		MaxWait:      getEnvDuration("KAFKA_MAX_WAIT", 500*time.Millisecond),
		ReadTimeout:  getEnvDuration("KAFKA_READ_TIMEOUT", 5*time.Second),
		ProcessDelay: getEnvDuration("PROCESS_DELAY", 0),
	}
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			part = trimSpaces(part)
			if part != "" {
				out = append(out, part)
			}
			start = i + 1
		}
	}
	return out
}

func trimSpaces(s string) string {
	// trim simples sem strings pkg pra manter compacto
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func main() {
	cfg := loadConfig()

	// Context cancelável p/ shutdown gracioso
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Captura SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Sinal recebido, encerrando consumer...")
		cancel()
	}()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID, // <- isso coloca em consumer group
		MinBytes: cfg.MinBytes,
		MaxBytes: cfg.MaxBytes,
		MaxWait:  cfg.MaxWait,
		// Se você quiser controlar commits manualmente:
		// CommitInterval: 0,
	})

	defer func() {
		_ = reader.Close()
	}()

	log.Printf("Consumer iniciado | brokers=%v topic=%s group=%s\n", cfg.Brokers, cfg.Topic, cfg.GroupID)

	for {
		// Evita ficar preso para sempre em ReadMessage
		readCtx, cancelRead := context.WithTimeout(ctx, cfg.ReadTimeout)
		msg, err := reader.ReadMessage(readCtx)
		cancelRead()

		if err != nil {
			// Se estamos encerrando, sai limpo
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				if ctx.Err() != nil {
					log.Println("Consumer finalizado.")
					return
				}
				// timeout sem cancelamento: só continua
				continue
			}
			log.Printf("Erro ao ler mensagem: %v\n", err)
			continue
		}

		// Processa
		if err := handleMessage(ctx, msg, cfg); err != nil {
			// Aqui você decide sua estratégia: retry, DLQ, log, métricas etc.
			log.Printf("Falha ao processar (topic=%s partition=%d offset=%d): %v\n",
				msg.Topic, msg.Partition, msg.Offset, err)

			// Se você usar commit manual, não comite em falhas.
			// Com commit automático (padrão), o comportamento depende da lib/config.
			continue
		}

		log.Printf("OK | key=%s partition=%d offset=%d value=%s\n",
			string(msg.Key), msg.Partition, msg.Offset, string(msg.Value))

		// Commit manual (se CommitInterval:0):
		// if err := reader.CommitMessages(ctx, msg); err != nil {
		// 	log.Printf("Erro no commit: %v\n", err)
		// }
	}
}

func handleMessage(ctx context.Context, msg kafka.Message, cfg Config) error {
	// Simule sua regra de negócio aqui.
	// Ex: decodificar JSON, chamar banco, chamar API, etc.

	if cfg.ProcessDelay > 0 {
		select {
		case <-time.After(cfg.ProcessDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Exemplo simples de validação:
	if len(msg.Value) == 0 {
		return fmt.Errorf("mensagem vazia")
	}
	return nil
}