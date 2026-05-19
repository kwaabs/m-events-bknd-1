package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kwaabs/m-events/internal/models"
	"github.com/kwaabs/m-events/internal/repository"
)

type CustomerHandler struct {
	repo *repository.CustomerRepository
}

func NewCustomerHandler(repo *repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

// ─── GET /api/v1/customers ───────────────────────────────────────────────────
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	hasTamper := strings.ToLower(q.Get("has_tamper")) == "true"

	filter := repository.CustomerFilter{
		Search:             q.Get("search"),
		RegionCode:         q.Get("region_code"),
		RegionName:         q.Get("region_name"),
		DistrictCode:       q.Get("district_code"),
		DistrictName:       q.Get("district_name"),
		ServiceType:        q.Get("service_type"),
		ServiceClass:       q.Get("service_class"),
		ServicePointNumber: q.Get("service_point_number"),
		CustomerType:       q.Get("customer_type"),
		ContractStatus:     q.Get("contract_status"),
		CMSContractStatus:  q.Get("cms_contract_status"),
		MeterMake:          q.Get("meter_make"),
		MeterModel:         q.Get("meter_model"),
		AccountType:        q.Get("account_type"),
		HasTamper:          hasTamper,
		HasCoordinates:     strings.ToLower(q.Get("has_coordinates")) == "true",
		NoCoordinates:      strings.ToLower(q.Get("no_coordinates")) == "true",
		Page:               page,
		PageSize:           pageSize,
	}

	customers, total, err := h.repo.ListWithLatestEvent(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch customers", err)
		return
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 50
	}

	dtos := make([]models.CustomerWithEvent, 0, len(customers))
	for _, c := range customers {
		dtos = append(dtos, models.ToCustomerWithEvent(c))
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))

	writeJSON(w, http.StatusOK, models.PaginatedResponse[models.CustomerWithEvent]{
		Data:       dtos,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	})
}

// ─── GET /api/v1/customers/account/{accountNumber} ───────────────────────────
func (h *CustomerHandler) GetByAccount(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.PathValue("accountNumber")
	if accountNumber == "" {
		writeError(w, http.StatusBadRequest, "account number is required", nil)
		return
	}

	cr, err := h.repo.GetByAccountNumber(r.Context(), accountNumber)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch customer", err)
		return
	}
	if cr == nil {
		writeError(w, http.StatusNotFound, "customer not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, models.ToCustomerWithEvent(*cr))
}

// ─── GET /api/v1/customers/no-coordinates ────────────────────────────────────
// Shortcut for customers that cannot be plotted on the map. Accepts all the
// same filters as the main list endpoint except no_coordinates (always true).
func (h *CustomerHandler) NoCoordinates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	filter := repository.CustomerFilter{
		Search:            q.Get("search"),
		RegionCode:        q.Get("region_code"),
		RegionName:        q.Get("region_name"),
		DistrictCode:      q.Get("district_code"),
		DistrictName:      q.Get("district_name"),
		ServiceType:       q.Get("service_type"),
		ServiceClass:      q.Get("service_class"),
		CustomerType:      q.Get("customer_type"),
		ContractStatus:    q.Get("contract_status"),
		CMSContractStatus: q.Get("cms_contract_status"),
		MeterMake:         q.Get("meter_make"),
		MeterModel:        q.Get("meter_model"),
		AccountType:       q.Get("account_type"),
		HasTamper:         strings.ToLower(q.Get("has_tamper")) == "true",
		NoCoordinates:     true, // always forced on this route
		Page:              page,
		PageSize:          pageSize,
	}

	customers, total, err := h.repo.ListWithLatestEvent(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch customers", err)
		return
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 50
	}

	dtos := make([]models.CustomerWithEvent, 0, len(customers))
	for _, c := range customers {
		dtos = append(dtos, models.ToCustomerWithEvent(c))
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))

	writeJSON(w, http.StatusOK, models.PaginatedResponse[models.CustomerWithEvent]{
		Data:       dtos,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	})
}

// ─── GET /api/v1/customers/meter/{meterNumber} ───────────────────────────────
func (h *CustomerHandler) GetByMeter(w http.ResponseWriter, r *http.Request) {
	meterNumber := r.PathValue("meterNumber")
	if meterNumber == "" {
		writeError(w, http.StatusBadRequest, "meter number is required", nil)
		return
	}

	cr, err := h.repo.GetByMeterNumber(r.Context(), meterNumber)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch customer", err)
		return
	}
	if cr == nil {
		writeError(w, http.StatusNotFound, "customer not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, models.ToCustomerWithEvent(*cr))
}

// ─── GET /api/v1/customers/meter/{meterNumber}/events ────────────────────────
// Query params: page, page_size, event_desc, event_code, from, to
func (h *CustomerHandler) EventsByMeter(w http.ResponseWriter, r *http.Request) {
	meterNumber := r.PathValue("meterNumber")
	if meterNumber == "" {
		writeError(w, http.StatusBadRequest, "meter number is required", nil)
		return
	}

	f, err := parseEventFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	events, total, err := h.repo.TamperEventsByMeter(r.Context(), meterNumber, f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch events", err)
		return
	}

	writeJSON(w, http.StatusOK, eventResponse(meterNumber, "", events, total, f))
}

// ─── GET /api/v1/customers/account/{accountNumber}/events ────────────────────
// Query params: page, page_size, event_desc, event_code, from, to
func (h *CustomerHandler) EventsByAccount(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.PathValue("accountNumber")
	if accountNumber == "" {
		writeError(w, http.StatusBadRequest, "account number is required", nil)
		return
	}

	f, err := parseEventFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	events, total, meterNumber, err := h.repo.TamperEventsByAccount(r.Context(), accountNumber, f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch events", err)
		return
	}
	if meterNumber == "" {
		writeError(w, http.StatusNotFound, "account not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, eventResponse(meterNumber, accountNumber, events, total, f))
}

// ─── GET /api/v1/health ──────────────────────────────────────────────────────
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// parseEventFilter reads all event filter params from the request query string.
// Returns an error if from/to are present but cannot be parsed.
func parseEventFilter(r *http.Request) (repository.EventFilter, error) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	f := repository.EventFilter{
		Page:      page,
		PageSize:  pageSize,
		EventDesc: q.Get("event_desc"),
		EventCode: q.Get("event_code"),
	}

	// Accept ISO 8601 datetime: 2026-04-01T00:00:00Z or 2026-04-01
	if raw := q.Get("from"); raw != "" {
		t, err := parseQueryTime(raw)
		if err != nil {
			return f, fmt.Errorf("invalid 'from' value %q — use ISO 8601, e.g. 2026-04-01T00:00:00Z", raw)
		}
		f.From = &t
	}
	if raw := q.Get("to"); raw != "" {
		t, err := parseQueryTime(raw)
		if err != nil {
			return f, fmt.Errorf("invalid 'to' value %q — use ISO 8601, e.g. 2026-04-30T23:59:59Z", raw)
		}
		f.To = &t
	}

	return f, nil
}

// parseQueryTime tries RFC3339 then date-only formats.
func parseQueryTime(s string) (time.Time, error) {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse %q", s)
}

type eventsEnvelope struct {
	MeterNumber   string               `json:"meter_number"`
	AccountNumber string               `json:"account_number,omitempty"`
	Total         int                  `json:"total"`
	Page          int                  `json:"page"`
	PageSize      int                  `json:"page_size"`
	TotalPages    int                  `json:"total_pages"`
	Events        []models.TamperEvent `json:"events"`
}

func eventResponse(
	meterNumber, accountNumber string,
	events []models.TamperEvent,
	total int,
	f repository.EventFilter,
) eventsEnvelope {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 50
	}
	if events == nil {
		events = []models.TamperEvent{}
	}
	return eventsEnvelope{
		MeterNumber:   meterNumber,
		AccountNumber: accountNumber,
		Total:         total,
		Page:          f.Page,
		PageSize:      f.PageSize,
		TotalPages:    int(math.Ceil(float64(total) / float64(f.PageSize))),
		Events:        events,
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	resp := errorResponse{Error: msg}
	if err != nil {
		resp.Message = err.Error()
	}
	writeJSON(w, status, resp)
}
