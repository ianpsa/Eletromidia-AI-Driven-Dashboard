package writer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sync"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
)

type ageRow struct {
	ID      string  `bigquery:"id"`
	X1819   float64 `bigquery:"x18_19"`
	X2029   float64 `bigquery:"x20_29"`
	X3039   float64 `bigquery:"x30_39"`
	X4049   float64 `bigquery:"x40_49"`
	X5059   float64 `bigquery:"x50_59"`
	X6069   float64 `bigquery:"x60_69"`
	X7079   float64 `bigquery:"x70_79"`
	X80Plus float64 `bigquery:"x80_plus"`
}

type genderRow struct {
	ID        string  `bigquery:"id"`
	Feminine  float64 `bigquery:"feminine"`
	Masculine float64 `bigquery:"masculine"`
}

type socialClassRow struct {
	ID      string  `bigquery:"id"`
	AClass  float64 `bigquery:"a_class"`
	B1Class float64 `bigquery:"b1_class"`
	B2Class float64 `bigquery:"b2_class"`
	C1Class float64 `bigquery:"c1_class"`
	C2Class float64 `bigquery:"c2_class"`
	DEClass float64 `bigquery:"de_class"`
}

type targetRow struct {
	ID            string `bigquery:"id"`
	AgeID         string `bigquery:"age_id"`
	GenderID      string `bigquery:"gender_id"`
	SocialClassID string `bigquery:"social_class_id"`
}

type geodataRow struct {
	ID             string  `bigquery:"id"`
	ImpressionHour int64   `bigquery:"impression_hour"`
	LocationID     int64   `bigquery:"location_id"`
	Uniques        float64 `bigquery:"uniques"`
	Latitude       string  `bigquery:"latitude"`
	Longitude      string  `bigquery:"longitude"`
	UfEstado       string  `bigquery:"uf_estado"`
	Cidade         string  `bigquery:"cidade"`
	Endereco       string  `bigquery:"endereco"`
	Numero         int64   `bigquery:"numero"`
	TargetID       string  `bigquery:"target_id"`
}

type KafkaEvent struct {
	ImpressionHour int64   `json:"impression_hour"`
	LocationID     int64   `json:"location_id"`
	Uniques        float64 `json:"uniques"`
	Latitude       string  `json:"latitude"`
	Longitude      string  `json:"longitude"`
	UfEstado       string  `json:"uf_estado"`
	Cidade         string  `json:"cidade"`
	Endereco       string  `json:"endereco"`
	Numero         int64   `json:"numero"`
	Target         map[string]map[string]float64  `json:"target"`
}

type TargetData struct {
	Idade        map[string]float64 `json:"idade"`
	Genero       map[string]float64 `json:"genero"`
	ClasseSocial map[string]float64 `json:"classe_social"`
}

type bqInserters struct {
	age               *bigquery.Inserter
	ageSchema         bigquery.Schema
	gender            *bigquery.Inserter
	genderSchema      bigquery.Schema
	socialClass       *bigquery.Inserter
	socialClassSchema bigquery.Schema
	target            *bigquery.Inserter
	targetSchema      bigquery.Schema
	geodata           *bigquery.Inserter
	geodataSchema     bigquery.Schema
}

func initInserters(ds *bigquery.Dataset) (*bqInserters, error) {
	ins := &bqInserters{}
	var err error

	if ins.ageSchema, err = bigquery.InferSchema(ageRow{}); err != nil {
		return nil, fmt.Errorf("InferSchema age: %w", err)
	}
	if ins.genderSchema, err = bigquery.InferSchema(genderRow{}); err != nil {
		return nil, fmt.Errorf("InferSchema gender: %w", err)
	}
	if ins.socialClassSchema, err = bigquery.InferSchema(socialClassRow{}); err != nil {
		return nil, fmt.Errorf("InferSchema social_class: %w", err)
	}
	if ins.targetSchema, err = bigquery.InferSchema(targetRow{}); err != nil {
		return nil, fmt.Errorf("InferSchema target: %w", err)
	}
	if ins.geodataSchema, err = bigquery.InferSchema(geodataRow{}); err != nil {
		return nil, fmt.Errorf("InferSchema geodata: %w", err)
	}

	ins.age = ds.Table("age").Inserter()
	ins.gender = ds.Table("gender").Inserter()
	ins.socialClass = ds.Table("social_class").Inserter()
	ins.target = ds.Table("target").Inserter()
	ins.geodata = ds.Table("geodata").Inserter()

	return ins, nil
}

var bqNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

func deterministicID(topic string, partition int, offset int64, table string) string {
	seed := fmt.Sprintf("%s:%d:%d:%s", topic, partition, offset, table)
	return uuid.NewSHA1(bqNamespace, []byte(seed)).String()
}

var singleQuoteRe = regexp.MustCompile(`'([^']*)'`)

func normalizeTarget(raw string) string {
	return singleQuoteRe.ReplaceAllString(raw, `"$1"`)
}

func validateMapKeys(td TargetData) error {
	for _, k := range []string{"18-19", "20-29", "30-39", "40-49", "50-59", "60-69", "70-79", "80+"} {
		if _, ok := td.Idade[k]; !ok {
			return fmt.Errorf("missing idade key %q", k)
		}
	}
	for _, k := range []string{"F", "M"} {
		if _, ok := td.Genero[k]; !ok {
			return fmt.Errorf("missing genero key %q", k)
		}
	}
	for _, k := range []string{"A", "B1", "B2", "C1", "C2", "DE"} {
		if _, ok := td.ClasseSocial[k]; !ok {
			return fmt.Errorf("missing classe_social key %q", k)
		}
	}
	return nil
}

type BufferedMessage struct {
	Topic     string
	Partition int
	Offset    int64
	Value     []byte
}

type Writer struct {
	ins     *bqInserters
	client  *bigquery.Client
	dataset string
	flushSz int

	mu     sync.Mutex
	buffer []BufferedMessage
}

func NewWriter(ctx context.Context, projectID, datasetID string, flushSize int) (*Writer, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %w", err)
	}

	ins, err := initInserters(client.Dataset(datasetID))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("initInserters: %w", err)
	}

	return &Writer{
		ins:     ins,
		client:  client,
		dataset: datasetID,
		flushSz: flushSize,
	}, nil
}

func (w *Writer) Add(msg BufferedMessage) bool {
	w.mu.Lock()
	w.buffer = append(w.buffer, msg)
	full := len(w.buffer) >= w.flushSz
	w.mu.Unlock()
	return full
}

func (w *Writer) Flush(ctx context.Context) error {
	w.mu.Lock()
	if len(w.buffer) == 0 {
		w.mu.Unlock()
		return nil
	}
	batch := w.buffer
	w.buffer = nil
	w.mu.Unlock()

	var (
		ageSavers         []*bigquery.StructSaver
		genderSavers      []*bigquery.StructSaver
		socialClassSavers []*bigquery.StructSaver
		targetSavers      []*bigquery.StructSaver
		geodataSavers     []*bigquery.StructSaver
	)

	parsed := 0
	for _, msg := range batch {
		var event KafkaEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("flush: json unmarshal error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			continue
		}

		var td TargetData
		// if err := json.Unmarshal([]byte(normalizeTarget(event.Target)), &td); err != nil {
		// 	log.Printf("flush: target parse error | partition=%d offset=%d: %v",
		// 		msg.Partition, msg.Offset, err)
		// 	continue
		// }
		td = TargetData{
			Idade: event.Target["idade"],
			Genero: event.Target["genero"],
			ClasseSocial: event.Target["classe_social"],
		}

		if err := validateMapKeys(td); err != nil {
			log.Printf("flush: target validation error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			continue
		}

		ageID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "age")
		genderID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "gender")
		socialClassID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "social_class")
		targetID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "target")
		geodataID := deterministicID(msg.Topic, msg.Partition, msg.Offset, "geodata")

		ageSavers = append(ageSavers, &bigquery.StructSaver{
			Schema: w.ins.ageSchema, InsertID: ageID,
			Struct: ageRow{
				ID: ageID, X1819: td.Idade["18-19"], X2029: td.Idade["20-29"],
				X3039: td.Idade["30-39"], X4049: td.Idade["40-49"],
				X5059: td.Idade["50-59"], X6069: td.Idade["60-69"],
				X7079: td.Idade["70-79"], X80Plus: td.Idade["80+"],
			},
		})

		genderSavers = append(genderSavers, &bigquery.StructSaver{
			Schema: w.ins.genderSchema, InsertID: genderID,
			Struct: genderRow{
				ID: genderID, Feminine: td.Genero["F"], Masculine: td.Genero["M"],
			},
		})

		socialClassSavers = append(socialClassSavers, &bigquery.StructSaver{
			Schema: w.ins.socialClassSchema, InsertID: socialClassID,
			Struct: socialClassRow{
				ID: socialClassID, AClass: td.ClasseSocial["A"],
				B1Class: td.ClasseSocial["B1"], B2Class: td.ClasseSocial["B2"],
				C1Class: td.ClasseSocial["C1"], C2Class: td.ClasseSocial["C2"],
				DEClass: td.ClasseSocial["DE"],
			},
		})

		targetSavers = append(targetSavers, &bigquery.StructSaver{
			Schema: w.ins.targetSchema, InsertID: targetID,
			Struct: targetRow{
				ID: targetID, AgeID: ageID,
				GenderID: genderID, SocialClassID: socialClassID,
			},
		})

		geodataSavers = append(geodataSavers, &bigquery.StructSaver{
			Schema: w.ins.geodataSchema, InsertID: geodataID,
			Struct: geodataRow{
				ID: geodataID, ImpressionHour: event.ImpressionHour,
				LocationID: event.LocationID, Uniques: event.Uniques,
				Latitude: event.Latitude, Longitude: event.Longitude,
				UfEstado: event.UfEstado, Cidade: event.Cidade,
				Endereco: event.Endereco, Numero: event.Numero,
				TargetID: targetID,
			},
		})

		parsed++
	}

	if parsed == 0 {
		return nil
	}

	type tableInsert struct {
		name     string
		inserter *bigquery.Inserter
		savers   []*bigquery.StructSaver
	}

	tables := []tableInsert{
		{"age", w.ins.age, ageSavers},
		{"gender", w.ins.gender, genderSavers},
		{"social_class", w.ins.socialClass, socialClassSavers},
		{"target", w.ins.target, targetSavers},
		{"geodata", w.ins.geodata, geodataSavers},
	}

	var insertErr error
	for _, t := range tables {
		if err := t.inserter.Put(ctx, t.savers); err != nil {
			log.Printf("flush: insert %s error (%d rows): %v", t.name, len(t.savers), err)
			insertErr = err
		} else {
			log.Printf("flush: %s inserted %d rows", t.name, len(t.savers))
		}
	}

	log.Printf("flush complete: %d/%d messages processed, 5 tables", parsed, len(batch))
	return insertErr
}

func (w *Writer) Pending() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.buffer)
}

func (w *Writer) Close() error {
	return w.client.Close()
}

func (w *Writer) HealthCheck(ctx context.Context) error {
	_, err := w.client.Dataset(w.dataset).Metadata(ctx)
	if err != nil {
		return fmt.Errorf("bigquery health check: %w", err)
	}
	return nil
}
