package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type FlushMetrics struct {
	// 1. Número de eventos enviados por flush.
	FlushEventCount *prometheus.HistogramVec

	// 2. Tempo entre o disparo do flush assíncrono e a confirmação do Bucket.
	FlushDuration *prometheus.HistogramVec

	// 4a. Lag atual por partição no momento do flush (valor mais recente).
	PartitionLag *prometheus.GaugeVec

	// 4b. Distribuição histórica do lag por partição ao longo dos flushes.
	PartitionLagHistogram *prometheus.HistogramVec

	// 5. Total acumulado de flushes disparados.
	FlushTotal *prometheus.CounterVec

	// 6. Total acumulado de erros ocorridos durante flushes.
	FlushErrorTotal *prometheus.CounterVec

	// 7. Janela de pŕocessamento do Flush (O quanto o flush processou comparado ao )
	FlushBatchSpread *prometheus.HistogramVec
}

func NewFlushMetrics(reg *prometheus.Registry) *FlushMetrics {

	// 1. Quantos eventos foram enviados por flush.
	flushEventCount := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_consumer_flush_event_count",
			Help:    "Número de eventos enviados por flush ao Bucket externo.",
			Buckets: []float64{1, 10, 50, 100, 200, 300, 400, 500},
		},
		[]string{"topic"},
	)

	// 2. Tempo total do flush assíncrono: do disparo até a confirmação do Bucket.
	flushDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_consumer_flush_duration_seconds",
			Help:    "Tempo entre o disparo do flush assíncrono e a confirmação do Bucket externo.",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"topic", "status"},
	)

	// 3/4a. Lag atual por partição: último offset disponível - último offset processado.
	partitionLag := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_partition_lag",
			Help: "Lag atual da partição no momento do flush (highWatermark - lastProcessedOffset).",
		},
		[]string{"topic", "partition"},
	)

	// 3/4b. Distribuição histórica do lag por partição ao longo dos flushes.
	partitionLagHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_consumer_partition_lag_histogram",
			Help:    "Distribuição histórica do lag por partição registrado a cada flush.",
			Buckets: []float64{0, 10, 50, 100, 500, 1000, 5000, 10000, 50000},
		},
		[]string{"topic", "partition"},
	)

	// 5. Total de flushes disparados — separado por reason para entender
	// a proporção entre flushes por capacidade e por tempo.
	flushTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_consumer_flush_total",
			Help: "Total acumulado de flushes disparados.",
		},
		[]string{"topic"},
	)

	// 6. Total de erros por flush — separado por tipo para facilitar
	// a distinção entre erros de rede, timeout, autenticação, etc.
	flushErrorTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_consumer_flush_error_total",
			Help: "Total acumulado de erros ocorridos durante flushes.",
		},
		[]string{"topic"},
	)

	reg.MustRegister(
		flushEventCount,
		flushDuration,
		partitionLag,
		partitionLagHistogram,
		flushTotal,
		flushErrorTotal,
	)

	return &FlushMetrics{
		FlushEventCount:       flushEventCount,
		FlushDuration:         flushDuration,
		PartitionLag:          partitionLag,
		PartitionLagHistogram: partitionLagHistogram,
		FlushTotal:            flushTotal,
		FlushErrorTotal:       flushErrorTotal,
	}
}
