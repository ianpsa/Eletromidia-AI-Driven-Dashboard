package service

import (
	"encoding/json"
	"fmt"
	"strings"
)

func sanitizeTarget(raw string) string {
	return strings.ReplaceAll(raw, "'", "\"")
}

type Target struct {
	Idade         map[string]float64
	Genero        map[string]float64
	Classe_Social map[string]float64
}

func ParseTarget(raw string) (Target, error) {
	sanitized := sanitizeTarget(raw)
	// fmt.Printf("Json tratado: %v \n", sanitized)

	var t Target
	err := json.Unmarshal([]byte(sanitized), &t)

	if err != nil {
		return Target{}, fmt.Errorf("[service/clean_target] Ocorreu um erro ao tratar o campo target: %v", err)
	}

	return t, err
}
