package writer

import (
	"bigquery-writer/internal/metrics"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	json "github.com/goccy/go-json"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/bigquery/storage/managedwriter"
)

type Writer struct {
	streams   *bqStreams
	bqClient  *bigquery.Client
	mwClient  *managedwriter.Client
	dataset   string
	projectID string
	flushSz   int

	mu     sync.Mutex
	buffer []BufferedMessage
}

func NewWriter(ctx context.Context, projectID, datasetID string, flushSize int) (*Writer, error) {
	bqClient, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %w", err)
	}

	mwClient, err := managedwriter.NewClient(ctx, projectID)
	if err != nil {
		if closeErr := bqClient.Close(); closeErr != nil {
			log.Printf("error closing bigquery client: %v", closeErr)
		}
		return nil, fmt.Errorf("managedwriter.NewClient: %w", err)
	}

	streams, err := initStreams(ctx, mwClient, projectID, datasetID)
	if err != nil {
		if closeErr := mwClient.Close(); closeErr != nil {
			log.Printf("error closing managedwriter client: %v", closeErr)
		}
		if closeErr := bqClient.Close(); closeErr != nil {
			log.Printf("error closing bigquery client: %v", closeErr)
		}
		return nil, fmt.Errorf("initStreams: %w", err)
	}

	return &Writer{
		streams:   streams,
		bqClient:  bqClient,
		mwClient:  mwClient,
		dataset:   datasetID,
		projectID: projectID,
		flushSz:   flushSize,
	}, nil
}

func (w *Writer) Add(msg BufferedMessage) bool {
	w.mu.Lock()
	w.buffer = append(w.buffer, msg)
	full := len(w.buffer) >= w.flushSz
	w.mu.Unlock()
	return full
}

func (w *Writer) Flush(ctx context.Context, m *metrics.FlushMetrics) error {
	w.mu.Lock()
	if len(w.buffer) == 0 {
		w.mu.Unlock()
		return nil
	}

	start := time.Now()

	batch := w.buffer
	w.buffer = nil
	w.mu.Unlock()

	msg := batch[0]
	topic := msg.Topic
	if topic == "" {
		topic = "unknown"
	}

	var (
		ageRows         [][]byte
		genderRows      [][]byte
		socialClassRows [][]byte
		targetRows      [][]byte
		geodataRows     [][]byte
	)

	type waterMarkSnapshots struct {
		highWatermark       int64
		lastProcessedOffset int64
		partition           int
		topic               string
	}

	snapshots := make(map[int32]waterMarkSnapshots)

	parsed := 0
	for _, msg := range batch {
		var event KafkaEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("flush: json unmarshal error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			m.FlushErrorTotal.WithLabelValues(topic).Inc()
			continue
		}

		current, exists := snapshots[int32(msg.Partition)]
		if !exists || msg.Offset > current.lastProcessedOffset {
			snapshots[int32(msg.Partition)] = waterMarkSnapshots{
				highWatermark:       msg.HighWatermark,
				lastProcessedOffset: msg.Offset,
				partition:           msg.Partition,
				topic:               msg.Topic,
			}
		}

		td := TargetData{
			Idade:        event.Target["idade"],
			Genero:       event.Target["genero"],
			ClasseSocial: event.Target["classe_social"],
		}

		if err := validateMapKeys(td); err != nil {
			log.Printf("flush: target validation error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			m.FlushErrorTotal.WithLabelValues(topic).Inc()
			continue
		}

		ageID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "age")
		genderID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "gender")
		socialClassID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "social_class")
		targetID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "target")
		geodataID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "geodata")

		ageBytes, err := encodeRow(w.streams.age.msgDesc, ageRowToData(ageRow{
			ID: ageID, X1819: td.Idade["18-19"], X2029: td.Idade["20-29"],
			X3039: td.Idade["30-39"], X4049: td.Idade["40-49"],
			X5059: td.Idade["50-59"], X6069: td.Idade["60-69"],
			X7079: td.Idade["70-79"], X80Plus: td.Idade["80+"],
		}))
		if err != nil {
			log.Printf("flush: encode age error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			m.FlushErrorTotal.WithLabelValues(topic).Inc()
			continue
		}

		genderBytes, err := encodeRow(w.streams.gender.msgDesc, genderRowToData(genderRow{
			ID: genderID, Feminine: td.Genero["F"], Masculine: td.Genero["M"],
		}))
		if err != nil {
			log.Printf("flush: encode gender error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			m.FlushErrorTotal.WithLabelValues(topic).Inc()
			continue
		}

		scBytes, err := encodeRow(w.streams.socialClass.msgDesc, socialClassRowToData(socialClassRow{
			ID: socialClassID, AClass: td.ClasseSocial["A"],
			B1Class: td.ClasseSocial["B1"], B2Class: td.ClasseSocial["B2"],
			C1Class: td.ClasseSocial["C1"], C2Class: td.ClasseSocial["C2"],
			DEClass: td.ClasseSocial["DE"],
		}))
		if err != nil {
			log.Printf("flush: encode social_class error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			m.FlushErrorTotal.WithLabelValues(topic).Inc()
			continue
		}

		tgtBytes, err := encodeRow(w.streams.target.msgDesc, targetRowToData(targetRow{
			ID: targetID, AgeID: ageID,
			GenderID: genderID, SocialClassID: socialClassID,
		}))
		if err != nil {
			log.Printf("flush: encode target error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			m.FlushErrorTotal.WithLabelValues(topic).Inc()
			continue
		}

		geoBytes, err := encodeRow(w.streams.geodata.msgDesc, geodataRowToData(geodataRow{
			ID: geodataID, ImpressionHour: event.ImpressionHour,
			LocationID: event.LocationID, Uniques: event.Uniques,
			Latitude: event.Latitude, Longitude: event.Longitude,
			UfEstado: event.UfEstado, Cidade: event.Cidade,
			Endereco: event.Endereco, Numero: event.Numero,
			TargetID: targetID,
		}))
		if err != nil {
			log.Printf("flush: encode geodata error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			m.FlushErrorTotal.WithLabelValues(topic).Inc()
			continue
		}

		ageRows = append(ageRows, ageBytes)
		genderRows = append(genderRows, genderBytes)
		socialClassRows = append(socialClassRows, scBytes)
		targetRows = append(targetRows, tgtBytes)
		geodataRows = append(geodataRows, geoBytes)

		parsed++
	}

	if parsed == 0 {
		return nil
	}

	for _, snap := range snapshots {
		lag := float64(snap.highWatermark - snap.lastProcessedOffset)
		m.PartitionLag.WithLabelValues(snap.topic, fmt.Sprintf("%d", snap.partition)).Set(lag)
		m.PartitionLagHistogram.WithLabelValues(snap.topic, fmt.Sprintf("%d", snap.partition)).Observe(lag)
	}

	type tableAppend struct {
		ts   *tableStream
		rows [][]byte
	}

	appends := []tableAppend{
		{w.streams.age, ageRows},
		{w.streams.gender, genderRows},
		{w.streams.socialClass, socialClassRows},
		{w.streams.target, targetRows},
		{w.streams.geodata, geodataRows},
	}

	var (
		appendErrors []error
		errMu        sync.Mutex
		wg           sync.WaitGroup
	)

	for _, a := range appends {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tStart := time.Now()

			result, err := a.ts.stream.AppendRows(ctx, a.rows)
			if err != nil {
				errMu.Lock()
				appendErrors = append(appendErrors, fmt.Errorf("%s: %w", a.ts.name, err))
				errMu.Unlock()
				log.Printf("flush: append %s error (%d rows): %v", a.ts.name, len(a.rows), err)
				m.FlushErrorTotal.WithLabelValues(topic).Inc()
				return
			}

			_, err = result.GetResult(ctx)
			if err != nil {
				errMu.Lock()
				appendErrors = append(appendErrors, fmt.Errorf("%s result: %w", a.ts.name, err))
				errMu.Unlock()
				log.Printf("flush: append %s result error: %v", a.ts.name, err)
				m.FlushErrorTotal.WithLabelValues(topic).Inc()
				return
			}

			// log.Printf("flush: %s appended %d rows", a.ts.name, len(a.rows))
			m.AppendLatency.WithLabelValues(topic, a.ts.name).Observe(time.Since(tStart).Seconds())
			m.AppendRowsTotal.WithLabelValues(topic, a.ts.name).Add(float64(len(a.rows)))
		}()
	}

	wg.Wait()

	// log.Printf("flush complete: %d/%d messages processed, 5 tables", parsed, len(batch))

	duration := time.Since(start).Seconds()
	m.FlushDuration.WithLabelValues(topic, "success").Observe(duration)
	m.FlushEventCount.WithLabelValues(topic).Observe(float64(parsed))
	m.FlushTotal.WithLabelValues(topic).Inc()

	if len(appendErrors) > 0 {
		return errors.Join(appendErrors...)
	}
	return nil
}

func (w *Writer) Pending() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.buffer)
}

func (w *Writer) Close() error {
	var errs []error
	for _, ts := range w.streams.all() {
		if err := ts.stream.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close stream %s: %w", ts.name, err))
		}
	}
	if err := w.mwClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close managedwriter client: %w", err))
	}
	if err := w.bqClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close bigquery client: %w", err))
	}
	return errors.Join(errs...)
}

func (w *Writer) HealthCheck(ctx context.Context) error {
	_, err := w.bqClient.Dataset(w.dataset).Metadata(ctx)
	if err != nil {
		return fmt.Errorf("bigquery health check: %w", err)
	}
	return nil
}
