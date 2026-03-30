package handler

import (
	"bff-storage/internal/models"
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const queryTimeout = 30 * time.Second

func (h *Handler) GetGeoPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filters := parseGeoFilters(r)

	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	result, err := h.BigQuery.QueryGeoPoints(ctx, filters)
	if err != nil {
		log.Printf("error querying geo points: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query geo points")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetDemographics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filters := parseGeoFilters(r)

	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	result, err := h.BigQuery.QueryDemographics(ctx, filters)
	if err != nil {
		log.Printf("error querying demographics: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query demographics")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetFilterOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filters := parseGeoFilters(r)

	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	result, err := h.BigQuery.QueryFilterOptions(ctx, filters)
	if err != nil {
		log.Printf("error querying filter options: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query filter options")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		writeError(w, http.StatusBadRequest, "group_by parameter is required (genero, faixa_etaria, classe_social)")
		return
	}

	validGroups := map[string]bool{"genero": true, "faixa_etaria": true, "classe_social": true}
	if !validGroups[groupBy] {
		writeError(w, http.StatusBadRequest, "invalid group_by value: must be genero, faixa_etaria, or classe_social")
		return
	}

	filters := parseGeoFilters(r)

	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	result, err := h.BigQuery.QueryCompare(ctx, filters, groupBy)
	if err != nil {
		log.Printf("error querying compare: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to query compare")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func parseGeoFilters(r *http.Request) models.GeoFilters {
	q := r.URL.Query()

	limit := 5000
	if v := q.Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	return models.GeoFilters{
		UfEstado:     splitParam(q.Get("uf_estado")),
		Cidade:       splitParam(q.Get("cidade")),
		Endereco:     splitParam(q.Get("endereco")),
		Horario:      splitParam(q.Get("horario")),
		Genero:       splitParam(q.Get("genero")),
		FaixaEtaria:  splitParam(q.Get("faixa_etaria")),
		ClasseSocial: splitParam(q.Get("classe_social")),
		Limit:        limit,
	}
}

func splitParam(raw string) []string {
	if raw == "" {
		return nil
	}
	var result []string
	for _, v := range strings.Split(raw, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}
