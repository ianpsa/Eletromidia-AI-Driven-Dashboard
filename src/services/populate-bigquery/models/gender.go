package models

type Gender struct {
	ID        string  `bigquery:"id"`
	Feminine  float64 `bigquery:"feminine"`
	Masculine float64 `bigquery:"masculine"`
}
