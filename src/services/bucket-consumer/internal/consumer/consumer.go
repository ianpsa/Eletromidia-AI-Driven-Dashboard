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

	kafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	cfg    config.Config
	writer *storage.Writer
}

func New(cfg config.Config, writer *storage.Writer) *Consumer {
	return &Consumer{cfg: cfg, writer: writer}
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
		Brokers:  c.cfg.KafkaBrokers,
		Topic:    c.cfg.KafkaTopic,
		GroupID:  c.cfg.KafkaGroupID,
		MinBytes: c.cfg.KafkaMinBytes,
		MaxBytes: c.cfg.KafkaMaxBytes,
		MaxWait:  c.cfg.KafkaMaxWait,
	})
	defer reader.Close()

	log.Printf("consumer iniciado | brokers=%v topico=%s grupo=%s bucket=gs://%s/%s",
		c.cfg.KafkaBrokers, c.cfg.KafkaTopic, c.cfg.KafkaGroupID,
		c.cfg.GCSBucket, c.cfg.GCSBasePath)

	ticker := time.NewTicker(c.cfg.FlushInterval)
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
		readCtx, cancelRead := context.WithTimeout(ctx, c.cfg.KafkaReadTimeout)
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

		if c.cfg.ProcessDelay > 0 {
			select {
			case <-time.After(c.cfg.ProcessDelay):
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
