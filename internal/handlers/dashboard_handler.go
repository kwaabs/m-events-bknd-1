package handlers

import (
	"net/http"

	"github.com/kwaabs/m-events/internal/repository"
)

type DashboardHandler struct {
	repo *repository.DashboardRepository
}

func NewDashboardHandler(repo *repository.DashboardRepository) *DashboardHandler {
	return &DashboardHandler{repo: repo}
}

// GET /api/v1/dashboard/summary
// Top-level KPIs for the dashboard header cards.
func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.repo.GetSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch summary", err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// GET /api/v1/dashboard/districts
// Per-district breakdown — powers the district table and choropleth.
// Optional query param: region_code
func (h *DashboardHandler) Districts(w http.ResponseWriter, r *http.Request) {
	regionCode := r.URL.Query().Get("region_code")

	districts, err := h.repo.GetDistricts(r.Context(), regionCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch districts", err)
		return
	}
	writeJSON(w, http.StatusOK, districts)
}

// GET /api/v1/dashboard/regions
// Per-region rollup.
func (h *DashboardHandler) Regions(w http.ResponseWriter, r *http.Request) {
	regions, err := h.repo.GetRegions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch regions", err)
		return
	}
	writeJSON(w, http.StatusOK, regions)
}
