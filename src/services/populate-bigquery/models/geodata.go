package models

type Geodata struct {
    ImpressionHour          string      `bigquery:"impression_hour"`
    Uniques                 float64     `bigquery:"uniques"`
    LocationID              int         `bigquery:"location_id"`
    Latitude                string      `bigquery:"latitude"`
    Longitude               string      `bigquery:"longitude"`
    UFEstado                string      `bigquery:"uf_estado"`
    Cidade                  string      `bigquery:"cidade"`
    Endereco                string      `bigquery:"endereco"`
    Numero                  int         `bigquery:"numero"`
    TargetID                string      `bigquery:"target_id"`
}