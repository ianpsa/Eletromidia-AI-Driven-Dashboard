package consumer

import (
	"bucket-consumer/internal/config"
	"bucket-consumer/internal/storage"
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	Cfg    config.Config
	writer *storage.Writer
}

func New(cfg config.Config, writer *storage.Writer) *Consumer {
	return &Consumer{Cfg: cfg, writer: writer}
}

var healthOnce sync.Once

func touchHealthFile() {
	healthOnce.Do(func() {
		if err := os.WriteFile("/tmp/healthy", nil, 0600); err != nil {
			log.Printf("liveness: falha ao criar /tmp/healthy: %v", err)
		}
	})
}

func (c *Consumer) Run(ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.Cfg.KafkaBrokers,
		Topic:    c.Cfg.KafkaTopic,
		GroupID:  c.Cfg.KafkaGroupID,
		MinBytes: c.Cfg.KafkaMinBytes,
		MaxBytes: c.Cfg.KafkaMaxBytes,
		MaxWait:  c.Cfg.KafkaMaxWait,
	})
	defer reader.Close()

	log.Printf("consumer iniciado | brokers=%v topico=%s grupo=%s bucket=gs://%s/%s",
		c.Cfg.KafkaBrokers, c.Cfg.KafkaTopic, c.Cfg.KafkaGroupID,
		c.Cfg.GCSBucket, c.Cfg.GCSBasePath)

	ticker := time.NewTicker(c.Cfg.FlushInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.writer.Flush(ctx); err != nil {
					log.Printf("erro no flush periodico: %v", err)
				}
			}
		}
	}()

	for {
		readCtx, cancelRead := context.WithTimeout(ctx, c.Cfg.KafkaReadTimeout)
		msg, err := reader.ReadMessage(readCtx)
		cancelRead()

		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Println("consumer encerrado")
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			log.Printf("erro ao ler mensagem: %v", err)
			continue
		}

		if c.Cfg.ProcessDelay > 0 {
			select {
			case <-time.After(c.Cfg.ProcessDelay):
			case <-ctx.Done():
				log.Println("consumer encerrado")
				return
			}
		}

		shouldFlush, err := c.writer.Add(storage.Message{
			Topic:     msg.Topic,
			Partition: msg.Partition,
			Offset:    msg.Offset,
			Timestamp: msg.Time,
			Value:     msg.Value,
		})
		if err != nil {
			log.Printf("erro ao adicionar ao buffer | particao=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			continue
		}

		if shouldFlush {
			if err := c.writer.Flush(ctx); err != nil {
				log.Printf("erro no flush: %v", err)
			}
		}

		touchHealthFile()
		log.Printf("bufferizado | particao=%d offset=%d pendentes=%d",
			msg.Partition, msg.Offset, c.writer.Pending())
	}
}

func KafkaReadinessProbe(brokerAddress string, timeout time.Duration) error {
	conn, err := kafka.DialContext(
		context.Background(),
		"tcp",
		brokerAddress,
	)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka broker: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(timeout))

	_, err = conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("kafka broker unreachable: %w", err)
	}

	return nil
}