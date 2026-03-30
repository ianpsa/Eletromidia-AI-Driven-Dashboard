package writer

import (
	"fmt"

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
	ImpressionHour int64                         `json:"impression_hour"`
	LocationID     int64                         `json:"location_id"`
	Uniques        float64                       `json:"uniques"`
	Latitude       string                        `json:"latitude"`
	Longitude      string                        `json:"longitude"`
	UfEstado       string                        `json:"uf_estado"`
	Cidade         string                        `json:"cidade"`
	Endereco       string                        `json:"endereco"`
	Numero         int64                         `json:"numero"`
	Target         map[string]map[string]float64 `json:"target"`
}

type TargetData struct {
	Idade        map[string]float64 `json:"idade"`
	Genero       map[string]float64 `json:"genero"`
	ClasseSocial map[string]float64 `json:"classe_social"`
}

type BufferedMessage struct {
	Topic         string
	Partition     int
	Offset        int64
	Value         []byte
	HighWatermark int64
}

var bqNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

func deterministicID(topic string, partition int, offset int64, table string) string {
	seed := fmt.Sprintf("%s:%d:%d:%s", topic, partition, offset, table)
	return uuid.NewSHA1(bqNamespace, []byte(seed)).String()
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
