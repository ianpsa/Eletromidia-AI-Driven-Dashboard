package consumer

import (
	"bigquery-writer/internal/config"
	"bigquery-writer/internal/writer"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	Cfg config.Config
	w   *writer.Writer
}

func New(cfg config.Config, w *writer.Writer) *Consumer {
	return &Consumer{Cfg: cfg, w: w}
}

var healthOnce sync.Once

func touchHealthFile() {
	healthOnce.Do(func() {
		if err := os.WriteFile("/tmp/healthy", nil, 0600); err != nil {
			log.Printf("liveness: failed to create /tmp/healthy: %v", err)
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

	log.Printf("consumer started | brokers=%v topic=%s group=%s",
		c.Cfg.KafkaBrokers, c.Cfg.KafkaTopic, c.Cfg.KafkaGroupID)

	ticker := time.NewTicker(c.Cfg.FlushInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.w.Flush(ctx); err != nil {
					log.Printf("periodic flush error: %v", err)
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
				log.Println("consumer stopped")
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			log.Printf("read message error: %v", err)
			continue
		}

		shouldFlush := c.w.Add(writer.BufferedMessage{
			Topic:     msg.Topic,
			Partition: msg.Partition,
			Offset:    msg.Offset,
			Value:     msg.Value,
		})

		if shouldFlush {
			if err := c.w.Flush(ctx); err != nil {
				log.Printf("flush error: %v", err)
			}
		}

		touchHealthFile()
		log.Printf("buffered | partition=%d offset=%d pending=%d",
			msg.Partition, msg.Offset, c.w.Pending())
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
