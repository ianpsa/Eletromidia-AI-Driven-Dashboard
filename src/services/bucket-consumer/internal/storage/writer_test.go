package storage

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// --------------- Helper ---------------

func unmarshalBuffer(t *testing.T, w *Writer, idx int) map[string]interface{} {
	t.Helper()
	if idx >= len(w.buffer) {
		t.Fatalf("buffer index %d out of range (len=%d)", idx, len(w.buffer))
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(w.buffer[idx].data, &raw); err != nil {
		t.Fatalf("failed to unmarshal buffer[%d]: %v", idx, err)
	}
	return raw
}

func newTestWriter(flushSz int) *Writer {
	return &Writer{flushSz: flushSz}
}

func validMessage(value string) Message {
	return Message{
		Topic:     "geodata",
		Partition: 3,
		Offset:    42,
		Timestamp: time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		Value:     []byte(value),
	}
}

// --------------- Grupo 1: TestNormalizeTarget ---------------

func TestNormalizeTarget(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single-quoted keys become double-quoted",
			input:    "{'idade': 0.5}",
			expected: `{"idade": 0.5}`,
		},
		{
			name:     "multiple single-quoted pairs",
			input:    "{'a': 1, 'b': 2}",
			expected: `{"a": 1, "b": 2}`,
		},
		{
			name:     "already double-quoted unchanged",
			input:    `{"idade": 0.5}`,
			expected: `{"idade": 0.5}`,
		},
		{
			name:     "no quotes unchanged",
			input:    "{idade: 0.5}",
			expected: "{idade: 0.5}",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "nested single-quoted keys",
			input:    "{'key': {'inner': 1}}",
			expected: `{"key": {"inner": 1}}`,
		},
		{
			name:     "value with spaces in single quotes",
			input:    "{'cidade': 'Sao Paulo'}",
			expected: `{"cidade": "Sao Paulo"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeTarget(tc.input)
			if got != tc.expected {
				t.Fatalf("normalizeTarget(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// --------------- Grupo 2: TestAdd ---------------

func TestAdd(t *testing.T) {
	t.Run("target string with single quotes is normalized to map", func(t *testing.T) {
		w := newTestWriter(100)
		msg := validMessage(`{"target": "{'idade': {'18-19': 0.5}}", "foo": 1}`)

		full, err := w.Add(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if full {
			t.Fatalf("expected not full")
		}

		raw := unmarshalBuffer(t, w, 0)
		target := raw["target"]
		if _, ok := target.(string); ok {
			t.Fatalf("expected target to be parsed as map, got string: %v", target)
		}
		if _, ok := target.(map[string]interface{}); !ok {
			t.Fatalf("expected target to be map, got %T", target)
		}
	})

	t.Run("target already a map remains unchanged", func(t *testing.T) {
		w := newTestWriter(100)
		msg := validMessage(`{"target": {"idade": {"18-19": 0.5}}, "foo": 1}`)

		_, err := w.Add(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		raw := unmarshalBuffer(t, w, 0)
		target := raw["target"]
		if _, ok := target.(map[string]interface{}); !ok {
			t.Fatalf("expected target to remain as map, got %T", target)
		}
	})

	t.Run("JSON without target field adds metadata without panic", func(t *testing.T) {
		w := newTestWriter(100)
		msg := validMessage(`{"foo": "bar", "count": 42}`)

		_, err := w.Add(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		raw := unmarshalBuffer(t, w, 0)
		if raw["foo"] != "bar" {
			t.Fatalf("expected foo=bar, got %v", raw["foo"])
		}
		if _, exists := raw["target"]; exists {
			t.Fatalf("expected no target key, but found one")
		}
		if _, exists := raw["_kafka_partition"]; !exists {
			t.Fatalf("expected _kafka_partition metadata")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		w := newTestWriter(100)
		msg := validMessage(`not json at all`)

		_, err := w.Add(msg)
		if err == nil {
			t.Fatalf("expected error for invalid JSON, got nil")
		}
		if !strings.Contains(err.Error(), "erro ao decodificar mensagem kafka") {
			t.Fatalf("unexpected error message: %v", err)
		}
		if len(w.buffer) != 0 {
			t.Fatalf("expected empty buffer after error, got %d", len(w.buffer))
		}
	})

	t.Run("empty JSON object adds only metadata", func(t *testing.T) {
		w := newTestWriter(100)
		msg := validMessage(`{}`)

		_, err := w.Add(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		raw := unmarshalBuffer(t, w, 0)
		if _, exists := raw["_kafka_partition"]; !exists {
			t.Fatalf("expected _kafka_partition in empty JSON")
		}
		if _, exists := raw["_kafka_offset"]; !exists {
			t.Fatalf("expected _kafka_offset in empty JSON")
		}
		if _, exists := raw["_kafka_timestamp"]; !exists {
			t.Fatalf("expected _kafka_timestamp in empty JSON")
		}
	})

	t.Run("kafka metadata enrichment with correct values", func(t *testing.T) {
		w := newTestWriter(100)
		msg := Message{
			Topic:     "geodata",
			Partition: 3,
			Offset:    42,
			Timestamp: time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
			Value:     []byte(`{"data": 1}`),
		}

		_, err := w.Add(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		raw := unmarshalBuffer(t, w, 0)

		partition, ok := raw["_kafka_partition"].(float64)
		if !ok {
			t.Fatalf("expected _kafka_partition as float64, got %T", raw["_kafka_partition"])
		}
		if int(partition) != 3 {
			t.Fatalf("expected _kafka_partition=3, got %v", partition)
		}

		offset, ok := raw["_kafka_offset"].(float64)
		if !ok {
			t.Fatalf("expected _kafka_offset as float64, got %T", raw["_kafka_offset"])
		}
		if int64(offset) != 42 {
			t.Fatalf("expected _kafka_offset=42, got %v", offset)
		}

		ts, ok := raw["_kafka_timestamp"].(string)
		if !ok {
			t.Fatalf("expected _kafka_timestamp as string, got %T", raw["_kafka_timestamp"])
		}
		if ts != "2025-06-15T10:30:00Z" {
			t.Fatalf("expected _kafka_timestamp=2025-06-15T10:30:00Z, got %q", ts)
		}
	})

	t.Run("zero timestamp fallback for buffered record date", func(t *testing.T) {
		w := newTestWriter(100)
		msg := Message{
			Topic:     "geodata",
			Partition: 0,
			Offset:    1,
			Timestamp: time.Time{},
			Value:     []byte(`{"data": 1}`),
		}

		before := time.Now().UTC()
		_, err := w.Add(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		rec := w.buffer[0]
		if rec.date.IsZero() {
			t.Fatalf("expected non-zero buffered record date for zero timestamp")
		}
		if rec.date.Before(before) {
			t.Fatalf("expected buffered date >= test start time")
		}

		raw := unmarshalBuffer(t, w, 0)
		ts := raw["_kafka_timestamp"].(string)
		if ts != "0001-01-01T00:00:00Z" {
			t.Fatalf("expected _kafka_timestamp for zero time, got %q", ts)
		}
	})

	t.Run("buffer full at threshold", func(t *testing.T) {
		w := newTestWriter(2)

		full1, err := w.Add(validMessage(`{"a": 1}`))
		if err != nil {
			t.Fatalf("unexpected error on 1st add: %v", err)
		}
		if full1 {
			t.Fatalf("expected not full after 1st add")
		}

		full2, err := w.Add(validMessage(`{"a": 2}`))
		if err != nil {
			t.Fatalf("unexpected error on 2nd add: %v", err)
		}
		if !full2 {
			t.Fatalf("expected full after 2nd add (flushSz=2)")
		}
	})

	t.Run("buffer not full below threshold", func(t *testing.T) {
		w := newTestWriter(5)

		full, err := w.Add(validMessage(`{"a": 1}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if full {
			t.Fatalf("expected not full with flushSz=5 and 1 message")
		}
	})

	t.Run("flush size 1 triggers immediately", func(t *testing.T) {
		w := newTestWriter(1)

		full, err := w.Add(validMessage(`{"a": 1}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !full {
			t.Fatalf("expected full with flushSz=1")
		}
	})

	t.Run("target string that fails parse after normalization stays as string", func(t *testing.T) {
		w := newTestWriter(100)
		msg := validMessage(`{"target": "not valid json even after normalization"}`)

		_, err := w.Add(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		raw := unmarshalBuffer(t, w, 0)
		target, ok := raw["target"].(string)
		if !ok {
			t.Fatalf("expected target to remain as string, got %T", raw["target"])
		}
		if target != "not valid json even after normalization" {
			t.Fatalf("expected original string preserved, got %q", target)
		}
	})

	t.Run("buffered record fields are correctly populated", func(t *testing.T) {
		w := newTestWriter(100)
		ts := time.Date(2025, 3, 20, 14, 0, 0, 0, time.UTC)
		msg := Message{
			Topic:     "test-topic",
			Partition: 7,
			Offset:    999,
			Timestamp: ts,
			Value:     []byte(`{"x": 1}`),
		}

		_, err := w.Add(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		rec := w.buffer[0]
		if rec.topic != "test-topic" {
			t.Fatalf("expected topic=test-topic, got %q", rec.topic)
		}
		if rec.partition != 7 {
			t.Fatalf("expected partition=7, got %d", rec.partition)
		}
		if rec.offset != 999 {
			t.Fatalf("expected offset=999, got %d", rec.offset)
		}
		if !rec.date.Equal(ts.UTC()) {
			t.Fatalf("expected date=%v, got %v", ts.UTC(), rec.date)
		}
	})
}

// --------------- Grupo 3: TestPending ---------------

func TestPending(t *testing.T) {
	t.Run("empty buffer returns 0", func(t *testing.T) {
		w := newTestWriter(100)
		if p := w.Pending(); p != 0 {
			t.Fatalf("expected 0 pending, got %d", p)
		}
	})

	t.Run("after 3 adds returns 3", func(t *testing.T) {
		w := newTestWriter(100)
		for i := 0; i < 3; i++ {
			if _, err := w.Add(validMessage(`{"i": ` + string(rune('0'+i)) + `}`)); err != nil {
				t.Fatal(err)
			}
		}
		if p := w.Pending(); p != 3 {
			t.Fatalf("expected 3 pending, got %d", p)
		}
	})
}
