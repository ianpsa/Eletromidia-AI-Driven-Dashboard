package models

type GeoPoint struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Value     float64 `json:"value"`
}

type GeoFilters struct {
	UfEstado     string
	Cidade       string
	Endereco     string
	Genero       string
	FaixaEtaria  string
	ClasseSocial string
	Limit        int
}

type GeoPointsResponse struct {
	Count  int        `json:"count"`
	Points []GeoPoint `json:"points"`
}

type DemographicSummary struct {
	Count       int                `json:"count"`
	Age         AgeSummary         `json:"age"`
	Gender      GenderSummary      `json:"gender"`
	SocialClass SocialClassSummary `json:"social_class"`
}

type AgeSummary struct {
	X18_19  float64 `json:"x18_19"`
	X20_29  float64 `json:"x20_29"`
	X30_39  float64 `json:"x30_39"`
	X40_49  float64 `json:"x40_49"`
	X50_59  float64 `json:"x50_59"`
	X60_69  float64 `json:"x60_69"`
	X70_79  float64 `json:"x70_79"`
	X80Plus float64 `json:"x80_plus"`
}

type GenderSummary struct {
	Feminine  float64 `json:"feminine"`
	Masculine float64 `json:"masculine"`
}

type SocialClassSummary struct {
	AClass  float64 `json:"a_class"`
	B1Class float64 `json:"b1_class"`
	B2Class float64 `json:"b2_class"`
	C1Class float64 `json:"c1_class"`
	C2Class float64 `json:"c2_class"`
	DEClass float64 `json:"de_class"`
}

type CompareGroup struct {
	Label string `json:"label"`
	DemographicSummary
}

type CompareResponse struct {
	Groups []CompareGroup `json:"groups"`
}

type FilterOptions struct {
	Estados   []string `json:"estados"`
	Cidades   []string `json:"cidades"`
	Enderecos []string `json:"enderecos"`
}
