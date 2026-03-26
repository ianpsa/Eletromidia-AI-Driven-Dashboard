package handler

import (
	"bff-storage/internal/models"
	"context"
	"log"
	"net/http"
	"strconv"
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

func parseGeoFilters(r *http.Request) models.GeoFilters {
	q := r.URL.Query()

	limit := 5000
	if v := q.Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	return models.GeoFilters{
		UfEstado:     q.Get("uf_estado"),
		Cidade:       q.Get("cidade"),
		Endereco:     q.Get("endereco"),
		Genero:       q.Get("genero"),
		FaixaEtaria:  q.Get("faixa_etaria"),
		ClasseSocial: q.Get("classe_social"),
		Limit:        limit,
	}
}
