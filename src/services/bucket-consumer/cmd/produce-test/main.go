package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func main() {
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "geodata",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	event := map[string]interface{}{
		"impression_hour": "2026-02-26 18:00:00",
		"location_id":     "12345",
		"uniques":         "150.5",
		"latitude":        "-23.5505",
		"longitude":       "-46.6333",
		"uf_estado":       "SP",
		"cidade":          "São Paulo",
		"endereco":        "Av Paulista",
		"numero":          "1000",
		"target":          "{'idade': {'18-19': 0.08, '20-29': 0.22, '30-39': 0.25, '40-49': 0.18, '50-59': 0.12, '60-69': 0.08, '70-79': 0.05, '80+': 0.02}, 'genero': {'F': 0.45, 'M': 0.55}, 'classe_social': {'A': 0.10, 'B1': 0.15, 'B2': 0.25, 'C1': 0.20, 'C2': 0.18, 'DE': 0.12}}",
	}

	value, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("erro ao serializar evento: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("test-key"),
		Value: value,
	})
	if err != nil {
		log.Fatalf("erro ao enviar mensagem: %v", err)
	}

	fmt.Println("mensagem enviada com sucesso para o topico geodata")
}
