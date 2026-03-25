package writer

import (
	"sync"
	"testing"

	"github.com/google/uuid"
)

// --------------- Helper ---------------

func validTargetData() TargetData {
	return TargetData{
		Idade: map[string]float64{
			"18-19": 0.1, "20-29": 0.2, "30-39": 0.15, "40-49": 0.15,
			"50-59": 0.1, "60-69": 0.1, "70-79": 0.1, "80+": 0.1,
		},
		Genero: map[string]float64{
			"F": 0.55, "M": 0.45,
		},
		ClasseSocial: map[string]float64{
			"A": 0.05, "B1": 0.1, "B2": 0.15,
			"C1": 0.2, "C2": 0.25, "DE": 0.25,
		},
	}
}

// --------------- Grupo 1: TestDeterministicID ---------------

func TestDeterministicID(t *testing.T) {
	t.Run("same inputs produce same UUID", func(t *testing.T) {
		a := deterministicID("geodata", 0, 42, "age")
		b := deterministicID("geodata", 0, 42, "age")
		if a != b {
			t.Fatalf("expected same UUID, got %q and %q", a, b)
		}
	})

	t.Run("different topic changes UUID", func(t *testing.T) {
		a := deterministicID("geodata", 0, 42, "age")
		b := deterministicID("events", 0, 42, "age")
		if a == b {
			t.Fatalf("expected different UUIDs for different topics, got %q", a)
		}
	})

	t.Run("different partition changes UUID", func(t *testing.T) {
		a := deterministicID("geodata", 0, 42, "age")
		b := deterministicID("geodata", 1, 42, "age")
		if a == b {
			t.Fatalf("expected different UUIDs for different partitions, got %q", a)
		}
	})

	t.Run("different offset changes UUID", func(t *testing.T) {
		a := deterministicID("geodata", 0, 42, "age")
		b := deterministicID("geodata", 0, 43, "age")
		if a == b {
			t.Fatalf("expected different UUIDs for different offsets, got %q", a)
		}
	})

	t.Run("different table changes UUID", func(t *testing.T) {
		a := deterministicID("geodata", 0, 42, "age")
		b := deterministicID("geodata", 0, 42, "gender")
		if a == b {
			t.Fatalf("expected different UUIDs for different tables, got %q", a)
		}
	})

	t.Run("output is valid UUID v5", func(t *testing.T) {
		id := deterministicID("geodata", 0, 42, "age")
		parsed, err := uuid.Parse(id)
		if err != nil {
			t.Fatalf("expected valid UUID, got parse error: %v", err)
		}
		if parsed.Version() != 5 {
			t.Fatalf("expected UUID version 5, got %d", parsed.Version())
		}
	})

	t.Run("empty strings do not panic", func(t *testing.T) {
		id := deterministicID("", 0, 0, "")
		if _, err := uuid.Parse(id); err != nil {
			t.Fatalf("expected valid UUID for empty inputs, got error: %v", err)
		}
	})
}

// --------------- Grupo 2: TestNormalizeTarget ---------------

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

// --------------- Grupo 3: TestValidateMapKeys ---------------

func TestValidateMapKeys(t *testing.T) {
	t.Run("all keys present returns nil", func(t *testing.T) {
		td := validTargetData()
		if err := validateMapKeys(td); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("extra keys still valid", func(t *testing.T) {
		td := validTargetData()
		td.Idade["99+"] = 0.01
		if err := validateMapKeys(td); err != nil {
			t.Fatalf("expected nil error with extra keys, got %v", err)
		}
	})

	t.Run("zero values still valid", func(t *testing.T) {
		td := validTargetData()
		for k := range td.Idade {
			td.Idade[k] = 0.0
		}
		if err := validateMapKeys(td); err != nil {
			t.Fatalf("expected nil error with zero values, got %v", err)
		}
	})

	idadeMissing := []string{"18-19", "20-29", "30-39", "80+"}
	for _, key := range idadeMissing {
		t.Run("missing idade key "+key, func(t *testing.T) {
			td := validTargetData()
			delete(td.Idade, key)
			err := validateMapKeys(td)
			if err == nil {
				t.Fatalf("expected error for missing idade key %q, got nil", key)
			}
		})
	}

	t.Run("missing genero key F", func(t *testing.T) {
		td := validTargetData()
		delete(td.Genero, "F")
		err := validateMapKeys(td)
		if err == nil {
			t.Fatalf("expected error for missing genero key F, got nil")
		}
	})

	t.Run("missing genero key M", func(t *testing.T) {
		td := validTargetData()
		delete(td.Genero, "M")
		err := validateMapKeys(td)
		if err == nil {
			t.Fatalf("expected error for missing genero key M, got nil")
		}
	})

	t.Run("missing classe_social key A", func(t *testing.T) {
		td := validTargetData()
		delete(td.ClasseSocial, "A")
		err := validateMapKeys(td)
		if err == nil {
			t.Fatalf("expected error for missing classe_social key A, got nil")
		}
	})

	t.Run("missing classe_social key DE", func(t *testing.T) {
		td := validTargetData()
		delete(td.ClasseSocial, "DE")
		err := validateMapKeys(td)
		if err == nil {
			t.Fatalf("expected error for missing classe_social key DE, got nil")
		}
	})

	t.Run("nil idade map", func(t *testing.T) {
		td := validTargetData()
		td.Idade = nil
		err := validateMapKeys(td)
		if err == nil {
			t.Fatalf("expected error for nil idade map, got nil")
		}
	})

	t.Run("all maps empty", func(t *testing.T) {
		td := TargetData{
			Idade:        map[string]float64{},
			Genero:       map[string]float64{},
			ClasseSocial: map[string]float64{},
		}
		err := validateMapKeys(td)
		if err == nil {
			t.Fatalf("expected error for empty maps, got nil")
		}
	})
}

// --------------- Grupo 4: TestAddAndPending ---------------

func TestAddAndPending(t *testing.T) {
	makeMsg := func(offset int64) BufferedMessage {
		return BufferedMessage{
			Topic:     "geodata",
			Partition: 0,
			Offset:    offset,
			Value:     []byte(`{"test": true}`),
		}
	}

	t.Run("sequential add and pending tracking", func(t *testing.T) {
		w := &Writer{flushSz: 3}

		if p := w.Pending(); p != 0 {
			t.Fatalf("expected 0 pending, got %d", p)
		}

		if full := w.Add(makeMsg(1)); full {
			t.Fatalf("expected not full after 1st add")
		}
		if p := w.Pending(); p != 1 {
			t.Fatalf("expected 1 pending, got %d", p)
		}

		if full := w.Add(makeMsg(2)); full {
			t.Fatalf("expected not full after 2nd add")
		}
		if p := w.Pending(); p != 2 {
			t.Fatalf("expected 2 pending, got %d", p)
		}

		if full := w.Add(makeMsg(3)); !full {
			t.Fatalf("expected full after 3rd add (flushSz=3)")
		}
		if p := w.Pending(); p != 3 {
			t.Fatalf("expected 3 pending, got %d", p)
		}
	})

	t.Run("flush size 1 triggers immediately", func(t *testing.T) {
		w := &Writer{flushSz: 1}
		if full := w.Add(makeMsg(1)); !full {
			t.Fatalf("expected full after 1st add with flushSz=1")
		}
	})

	t.Run("concurrent adds are safe", func(t *testing.T) {
		w := &Writer{flushSz: 1000}
		var wg sync.WaitGroup
		n := 100
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(offset int64) {
				defer wg.Done()
				w.Add(makeMsg(offset))
			}(int64(i))
		}
		wg.Wait()
		if p := w.Pending(); p != n {
			t.Fatalf("expected %d pending after concurrent adds, got %d", n, p)
		}
	})
}
