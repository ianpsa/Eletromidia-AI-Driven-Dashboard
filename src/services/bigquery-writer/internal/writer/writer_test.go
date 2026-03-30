package writer

import (
	"sync"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
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

// --------------- Grupo 2: TestValidateMapKeys ---------------

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

// --------------- Grupo 5: TestDeriveProtoDescriptor ---------------

func TestDeriveProtoDescriptor(t *testing.T) {
	tests := []struct {
		name       string
		row        interface{}
		scope      string
		wantFields int
	}{
		{"age", ageRow{}, "age", 9},
		{"gender", genderRow{}, "gender", 3},
		{"social_class", socialClassRow{}, "social_class", 7},
		{"target", targetRow{}, "target", 4},
		{"geodata", geodataRow{}, "geodata", 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := bigquery.InferSchema(tt.row)
			if err != nil {
				t.Fatalf("InferSchema: %v", err)
			}

			dp, msgDesc, err := deriveProtoDescriptor(schema, tt.scope)
			if err != nil {
				t.Fatalf("deriveProtoDescriptor: %v", err)
			}

			if dp == nil {
				t.Fatal("expected non-nil DescriptorProto")
			}

			if got := msgDesc.Fields().Len(); got != tt.wantFields {
				t.Fatalf("expected %d fields, got %d", tt.wantFields, got)
			}
		})
	}
}

// --------------- Grupo 6: TestEncodeRow ---------------

func TestEncodeRow(t *testing.T) {
	schema, err := bigquery.InferSchema(genderRow{})
	if err != nil {
		t.Fatalf("InferSchema: %v", err)
	}

	_, msgDesc, err := deriveProtoDescriptor(schema, "gender")
	if err != nil {
		t.Fatalf("deriveProtoDescriptor: %v", err)
	}

	t.Run("round trip preserves values", func(t *testing.T) {
		data := rowData{
			"id":        "test-id-123",
			"feminine":  0.55,
			"masculine": 0.45,
		}

		encoded, err := encodeRow(msgDesc, data)
		if err != nil {
			t.Fatalf("encodeRow: %v", err)
		}

		if len(encoded) == 0 {
			t.Fatal("expected non-empty encoded bytes")
		}

		decoded := dynamicpb.NewMessage(msgDesc)
		if err := proto.Unmarshal(encoded, decoded); err != nil {
			t.Fatalf("proto.Unmarshal: %v", err)
		}

		fields := msgDesc.Fields()
		idField := fields.ByName("id")
		if got := decoded.Get(idField).String(); got != "test-id-123" {
			t.Fatalf("expected id 'test-id-123', got %q", got)
		}

		femField := fields.ByName("feminine")
		if got := decoded.Get(femField).Float(); got != 0.55 {
			t.Fatalf("expected feminine 0.55, got %f", got)
		}

		mascField := fields.ByName("masculine")
		if got := decoded.Get(mascField).Float(); got != 0.45 {
			t.Fatalf("expected masculine 0.45, got %f", got)
		}
	})

	t.Run("missing fields are skipped", func(t *testing.T) {
		data := rowData{"id": "partial"}

		encoded, err := encodeRow(msgDesc, data)
		if err != nil {
			t.Fatalf("encodeRow with partial data: %v", err)
		}

		decoded := dynamicpb.NewMessage(msgDesc)
		if err := proto.Unmarshal(encoded, decoded); err != nil {
			t.Fatalf("proto.Unmarshal: %v", err)
		}

		fields := msgDesc.Fields()
		if got := decoded.Get(fields.ByName("id")).String(); got != "partial" {
			t.Fatalf("expected id 'partial', got %q", got)
		}
		if got := decoded.Get(fields.ByName("feminine")).Float(); got != 0.0 {
			t.Fatalf("expected default 0.0 for feminine, got %f", got)
		}
	})
}

// --------------- Grupo 7: TestRowToData ---------------

func TestRowToData(t *testing.T) {
	t.Run("ageRowToData", func(t *testing.T) {
		row := ageRow{
			ID: "a1", X1819: 0.1, X2029: 0.2, X3039: 0.3, X4049: 0.4,
			X5059: 0.5, X6069: 0.6, X7079: 0.7, X80Plus: 0.8,
		}
		data := ageRowToData(row)
		if data["id"] != "a1" {
			t.Fatalf("expected id 'a1', got %v", data["id"])
		}
		if data["x18_19"] != 0.1 {
			t.Fatalf("expected x18_19 0.1, got %v", data["x18_19"])
		}
		if len(data) != 9 {
			t.Fatalf("expected 9 fields, got %d", len(data))
		}
	})

	t.Run("genderRowToData", func(t *testing.T) {
		row := genderRow{ID: "g1", Feminine: 0.55, Masculine: 0.45}
		data := genderRowToData(row)
		if data["feminine"] != 0.55 {
			t.Fatalf("expected feminine 0.55, got %v", data["feminine"])
		}
		if len(data) != 3 {
			t.Fatalf("expected 3 fields, got %d", len(data))
		}
	})

	t.Run("socialClassRowToData", func(t *testing.T) {
		row := socialClassRow{
			ID: "sc1", AClass: 0.05, B1Class: 0.1, B2Class: 0.15,
			C1Class: 0.2, C2Class: 0.25, DEClass: 0.25,
		}
		data := socialClassRowToData(row)
		if data["a_class"] != 0.05 {
			t.Fatalf("expected a_class 0.05, got %v", data["a_class"])
		}
		if len(data) != 7 {
			t.Fatalf("expected 7 fields, got %d", len(data))
		}
	})

	t.Run("targetRowToData", func(t *testing.T) {
		row := targetRow{ID: "t1", AgeID: "a1", GenderID: "g1", SocialClassID: "sc1"}
		data := targetRowToData(row)
		if data["age_id"] != "a1" {
			t.Fatalf("expected age_id 'a1', got %v", data["age_id"])
		}
		if len(data) != 4 {
			t.Fatalf("expected 4 fields, got %d", len(data))
		}
	})

	t.Run("geodataRowToData", func(t *testing.T) {
		row := geodataRow{
			ID: "geo1", ImpressionHour: 14, LocationID: 42,
			Uniques: 1.5, Latitude: "-23.5", Longitude: "-46.6",
			UfEstado: "SP", Cidade: "SaoPaulo", Endereco: "Rua X",
			Numero: 100, TargetID: "t1",
		}
		data := geodataRowToData(row)
		if data["impression_hour"] != int64(14) {
			t.Fatalf("expected impression_hour 14, got %v", data["impression_hour"])
		}
		if data["latitude"] != "-23.5" {
			t.Fatalf("expected latitude '-23.5', got %v", data["latitude"])
		}
		if len(data) != 11 {
			t.Fatalf("expected 11 fields, got %d", len(data))
		}
	})
}

// --------------- Grupo 8: TestEncodeRowAllTables ---------------

func TestEncodeRowAllTables(t *testing.T) {
	tables := []struct {
		name string
		row  interface{}
		data func() rowData
	}{
		{"age", ageRow{}, func() rowData {
			return ageRowToData(ageRow{
				ID: "a1", X1819: 0.1, X2029: 0.2, X3039: 0.3, X4049: 0.4,
				X5059: 0.5, X6069: 0.6, X7079: 0.7, X80Plus: 0.8,
			})
		}},
		{"geodata", geodataRow{}, func() rowData {
			return geodataRowToData(geodataRow{
				ID: "geo1", ImpressionHour: 14, LocationID: 42,
				Uniques: 1.5, Latitude: "-23.5", Longitude: "-46.6",
				UfEstado: "SP", Cidade: "SaoPaulo", Endereco: "Rua X",
				Numero: 100, TargetID: "t1",
			})
		}},
	}

	for _, tt := range tables {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := bigquery.InferSchema(tt.row)
			if err != nil {
				t.Fatalf("InferSchema: %v", err)
			}

			_, msgDesc, err := deriveProtoDescriptor(schema, tt.name)
			if err != nil {
				t.Fatalf("deriveProtoDescriptor: %v", err)
			}

			encoded, err := encodeRow(msgDesc, tt.data())
			if err != nil {
				t.Fatalf("encodeRow: %v", err)
			}

			if len(encoded) == 0 {
				t.Fatal("expected non-empty encoded bytes")
			}

			decoded := dynamicpb.NewMessage(msgDesc)
			if err := proto.Unmarshal(encoded, decoded); err != nil {
				t.Fatalf("proto.Unmarshal round-trip failed: %v", err)
			}
		})
	}
}
