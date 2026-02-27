package models

type SocialClass struct {
    ID           string     `bigquery:"id"`
    A_Class      float64    `bigquery:"a_class"`
    B1_Class     float64    `bigquery:"b1_class"`
    B2_Class     float64    `bigquery:"b2_class"`
    C1_Class     float64    `bigquery:"c1_class"`
    C2_Class     float64    `bigquery:"c2_class"`
    DE_Class     float64    `bigquery:"de_class"`
}
