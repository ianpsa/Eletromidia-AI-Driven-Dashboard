package models

type Age struct {
	ID       string  `bigquery:"id"`
	X18_19   float64 `bigquery:"x18_19"`
	X20_29   float64 `bigquery:"x20_29"`
	X30_39   float64 `bigquery:"x30_39"`
	X40_49   float64 `bigquery:"x40_49"`
	X50_59   float64 `bigquery:"x50_59"`
	X60_69   float64 `bigquery:"x60_69"`
	X70_79   float64 `bigquery:"x70_79"`
	X80_plus float64 `bigquery:"x80_plus"`
}
