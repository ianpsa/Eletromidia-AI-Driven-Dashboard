package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

// ─── Config ──────────────────────────────────────────────────────────────────

type Config struct {
	KafkaBrokers  string
	KafkaTopic    string
	KafkaGroupID  string
	KafkaMinBytes int
	KafkaMaxBytes int
	GCPProjectID  string
	BQDatasetID   string
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func loadConfig() Config {
	return Config{
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "my-cluster-kafka-bootstrap.kafka.svc:9092"),
		KafkaTopic:    getEnv("KAFKA_TOPIC", "geodata"),
		KafkaGroupID:  getEnv("KAFKA_GROUP_ID", "bigquery-writer-group"),
		KafkaMinBytes: getEnvInt("KAFKA_MIN_BYTES", 1000),
		KafkaMaxBytes: getEnvInt("KAFKA_MAX_BYTES", 10_000_000),
		GCPProjectID:  getEnv("GCP_PROJECT_ID", ""),
		BQDatasetID:   getEnv("BQ_DATASET_ID", ""),
	}
}

func validateConfig(cfg Config) error {
	var missing []string
	if cfg.GCPProjectID == "" {
		missing = append(missing, "GCP_PROJECT_ID")
	}
	if cfg.BQDatasetID == "" {
		missing = append(missing, "BQ_DATASET_ID")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ─── Kafka Message Structs ────────────────────────────────────────────────────

type KafkaEvent struct {
	ImpressionHour string `json:"impression_hour"`
	LocationID     string `json:"location_id"`
	Uniques        string `json:"uniques"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	UfEstado       string `json:"uf_estado"`
	Cidade         string `json:"cidade"`
	Endereco       string `json:"endereco"`
	Numero         string `json:"numero"`
	Target         string `json:"target"`
}

type TargetData struct {
	Idade        map[string]float64 `json:"idade"`
	Genero       map[string]float64 `json:"genero"`
	ClasseSocial map[string]float64 `json:"classe_social"`
}

// ─── BigQuery Row Structs ─────────────────────────────────────────────────────

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
	ID            string  `bigquery:"id"`
	AgeID         *string `bigquery:"age_id"`
	GenderID      *string `bigquery:"gender_id"`
	SocialClassID *string `bigquery:"social_class_id"`
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
	TargetID       *string `bigquery:"target_id"`
}

// ─── BQ Inserters ─────────────────────────────────────────────────────────────

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

// ─── Deterministic IDs ────────────────────────────────────────────────────────

var bqNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

func deterministicID(msg kafka.Message, table string) string {
	seed := fmt.Sprintf("%s:%d:%d:%s", msg.Topic, msg.Partition, msg.Offset, table)
	return uuid.NewSHA1(bqNamespace, []byte(seed)).String()
}

// ─── BigQuery Insert Functions ────────────────────────────────────────────────

func insertAge(ctx context.Context, ins *bqInserters, msg kafka.Message, idade map[string]float64) (string, error) {
	id := deterministicID(msg, "age")
	row := ageRow{
		ID:      id,
		X1819:   idade["18-19"],
		X2029:   idade["20-29"],
		X3039:   idade["30-39"],
		X4049:   idade["40-49"],
		X5059:   idade["50-59"],
		X6069:   idade["60-69"],
		X7079:   idade["70-79"],
		X80Plus: idade["80+"],
	}
	saver := &bigquery.StructSaver{Schema: ins.ageSchema, InsertID: id, Struct: row}
	if err := ins.age.Put(ctx, saver); err != nil {
		return "", fmt.Errorf("insertAge: %w", err)
	}
	log.Printf("age inserted          | id=%s", id)
	return id, nil
}

func insertGender(ctx context.Context, ins *bqInserters, msg kafka.Message, genero map[string]float64) (string, error) {
	id := deterministicID(msg, "gender")
	row := genderRow{
		ID:        id,
		Feminine:  genero["F"],
		Masculine: genero["M"],
	}
	saver := &bigquery.StructSaver{Schema: ins.genderSchema, InsertID: id, Struct: row}
	if err := ins.gender.Put(ctx, saver); err != nil {
		return "", fmt.Errorf("insertGender: %w", err)
	}
	log.Printf("gender inserted       | id=%s", id)
	return id, nil
}

func insertSocialClass(ctx context.Context, ins *bqInserters, msg kafka.Message, classeSocial map[string]float64) (string, error) {
	id := deterministicID(msg, "social_class")
	row := socialClassRow{
		ID:      id,
		AClass:  classeSocial["A"],
		B1Class: classeSocial["B1"],
		B2Class: classeSocial["B2"],
		C1Class: classeSocial["C1"],
		C2Class: classeSocial["C2"],
		DEClass: classeSocial["DE"],
	}
	saver := &bigquery.StructSaver{Schema: ins.socialClassSchema, InsertID: id, Struct: row}
	if err := ins.socialClass.Put(ctx, saver); err != nil {
		return "", fmt.Errorf("insertSocialClass: %w", err)
	}
	log.Printf("social_class inserted | id=%s", id)
	return id, nil
}

func insertTarget(ctx context.Context, ins *bqInserters, msg kafka.Message, ageID, genderID, socialClassID string) (string, error) {
	id := deterministicID(msg, "target")
	row := targetRow{
		ID:            id,
		AgeID:         &ageID,
		GenderID:      &genderID,
		SocialClassID: &socialClassID,
	}
	saver := &bigquery.StructSaver{Schema: ins.targetSchema, InsertID: id, Struct: row}
	if err := ins.target.Put(ctx, saver); err != nil {
		return "", fmt.Errorf("insertTarget: %w", err)
	}
	log.Printf("target inserted       | id=%s", id)
	return id, nil
}

func insertGeodata(ctx context.Context, ins *bqInserters, msg kafka.Message, event KafkaEvent, targetID string) error {
	id := deterministicID(msg, "geodata")

	impressionHour, err := strconv.ParseInt(event.ImpressionHour, 10, 64)
	if err != nil {
		return fmt.Errorf("insertGeodata: invalid impressionHour %q: %w", event.ImpressionHour, err)
	}

	locationID, err := strconv.ParseInt(event.LocationID, 10, 64)
	if err != nil {
		return fmt.Errorf("insertGeodata: invalid location_id %q: %w", event.LocationID, err)
	}

	uniques, err := strconv.ParseFloat(event.Uniques, 64)
	if err != nil {
		return fmt.Errorf("insertGeodata: invalid uniques %q: %w", event.Uniques, err)
	}

	numero, err := strconv.ParseInt(event.Numero, 10, 64)
	if err != nil {
		return fmt.Errorf("insertGeodata: invalid numero %q: %w", event.Numero, err)
	}

	row := geodataRow{
		ID:             id,
		ImpressionHour: impressionHour,
		LocationID:     locationID,
		Uniques:        uniques,
		Latitude:       event.Latitude,
		Longitude:      event.Longitude,
		UfEstado:       event.UfEstado,
		Cidade:         event.Cidade,
		Endereco:       event.Endereco,
		Numero:         numero,
		TargetID:       &targetID,
	}
	saver := &bigquery.StructSaver{Schema: ins.geodataSchema, InsertID: id, Struct: row}
	if err := ins.geodata.Put(ctx, saver); err != nil {
		return fmt.Errorf("insertGeodata: %w", err)
	}
	log.Printf("geodata inserted      | id=%s target_id=%s", id, targetID)
	return nil
}

// ─── Validation ───────────────────────────────────────────────────────────────

func validateMapKeys(td TargetData) error {
	idadeKeys := []string{"18-19", "20-29", "30-39", "40-49", "50-59", "60-69", "70-79", "80+"}
	for _, k := range idadeKeys {
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

// ─── Message Handler ──────────────────────────────────────────────────────────

var singleQuoteRe = regexp.MustCompile(`'([^']*)'`)

func normalizeTarget(raw string) string {
	return singleQuoteRe.ReplaceAllString(raw, `"$1"`)
}

// handleMessage inserts rows into all 5 tables sequentially.
// On partial failure, the caller does NOT commit the Kafka offset,
// so the message will be redelivered. Deterministic UUID v5 IDs
// ensure BigQuery deduplicates already-inserted rows via InsertID.
func handleMessage(ctx context.Context, ins *bqInserters, msg kafka.Message) error {
	var event KafkaEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	var targetData TargetData
	if err := json.Unmarshal([]byte(normalizeTarget(event.Target)), &targetData); err != nil {
		return fmt.Errorf("target parse: %w", err)
	}

	if err := validateMapKeys(targetData); err != nil {
		return fmt.Errorf("target validation: %w", err)
	}

	ageID, err := insertAge(ctx, ins, msg, targetData.Idade)
	if err != nil {
		return err
	}

	genderID, err := insertGender(ctx, ins, msg, targetData.Genero)
	if err != nil {
		return err
	}

	socialClassID, err := insertSocialClass(ctx, ins, msg, targetData.ClasseSocial)
	if err != nil {
		return err
	}

	targetID, err := insertTarget(ctx, ins, msg, ageID, genderID, socialClassID)
	if err != nil {
		return err
	}

	return insertGeodata(ctx, ins, msg, event, targetID)
}

// ─── Liveness Probe ───────────────────────────────────────────────────────────

var healthOnce sync.Once

func touchHealthFile() {
	healthOnce.Do(func() {
		if err := os.WriteFile("/tmp/healthy", nil, 0600); err != nil {
			log.Printf("liveness: failed to create /tmp/healthy: %v", err)
		}
	})
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	cfg := loadConfig()

	if err := validateConfig(cfg); err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bqClient, err := bigquery.NewClient(ctx, cfg.GCPProjectID)
	if err != nil {
		log.Fatalf("bigquery.NewClient: %v", err)
	}
	defer bqClient.Close()

	ins, err := initInserters(bqClient.Dataset(cfg.BQDatasetID))
	if err != nil {
		log.Fatalf("initInserters: %v", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(cfg.KafkaBrokers, ","),
		Topic:    cfg.KafkaTopic,
		GroupID:  cfg.KafkaGroupID,
		MinBytes: cfg.KafkaMinBytes,
		MaxBytes: cfg.KafkaMaxBytes,
		MaxWait:  500 * time.Millisecond,
	})
	defer reader.Close()

	log.Printf("Consumer iniciado | brokers=%s topic=%s group=%s",
		cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutdown signal received")
		cancel()
	}()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Consumer stopped")
				return
			}
			log.Printf("FetchMessage error: %v", err)
			continue
		}

		if err := handleMessage(ctx, ins, msg); err != nil {
			log.Printf("handleMessage error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			continue // offset NOT committed — message will be redelivered
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("CommitMessages error | partition=%d offset=%d: %v",
				msg.Partition, msg.Offset, err)
			// message will be redelivered on next startup, deterministic UUIDs handle dedup
		}

		touchHealthFile()
		log.Printf("OK | key=%s partition=%d offset=%d",
			string(msg.Key), msg.Partition, msg.Offset)
	}
}
