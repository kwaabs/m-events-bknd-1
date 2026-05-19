package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type DashboardRepository struct {
	db *bun.DB
}

func NewDashboardRepository(db *bun.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// ── Response types ────────────────────────────────────────────────────────────

type DashboardSummary struct {
	TotalMeters              int64      `json:"total_meters"              bun:"total_meters"`
	MetersWithCoordinates    int64      `json:"meters_with_coordinates"   bun:"meters_with_coordinates"`
	MetersWithoutCoordinates int64      `json:"meters_without_coordinates" bun:"meters_without_coordinates"`
	CoordinateCoverage       float64    `json:"coordinate_coverage_pct"   bun:"coordinate_coverage_pct"`
	MetersWithEvents         int64      `json:"meters_with_events"        bun:"meters_with_events"`
	TotalEvents              int64      `json:"total_events"              bun:"total_events"`
	TotalDistricts           int64      `json:"total_districts"           bun:"total_districts"`
	TotalRegions             int64      `json:"total_regions"             bun:"total_regions"`
	PrepaidMeters            int64      `json:"prepaid_meters"            bun:"prepaid_meters"`
	PostpaidMeters           int64      `json:"postpaid_meters"           bun:"postpaid_meters"`
	ActiveContracts          int64      `json:"active_contracts"          bun:"active_contracts"`
	LatestEventTime          *time.Time `json:"latest_event_time"         bun:"latest_event_time"`
	LastRefreshed            *time.Time `json:"last_refreshed"            bun:"last_refreshed"`
}

type DistrictSummary struct {
	DistrictCode             string     `json:"district_code"              bun:"districtcode"`
	DistrictName             string     `json:"district_name"              bun:"districtname"`
	RegionCode               string     `json:"region_code"                bun:"regioncode"`
	RegionName               string     `json:"region_name"                bun:"regionname"`
	TotalMeters              int64      `json:"total_meters"               bun:"total_meters"`
	MetersWithEvents         int64      `json:"meters_with_events"         bun:"meters_with_events"`
	MetersWithCoordinates    int64      `json:"meters_with_coordinates"    bun:"meters_with_coordinates"`
	MetersWithoutCoordinates int64      `json:"meters_without_coordinates" bun:"meters_without_coordinates"`
	TotalEvents              int64      `json:"total_events"               bun:"total_events"`
	LatestEventTime          *time.Time `json:"latest_event_time"          bun:"latest_event_time"`
}

type RegionSummary struct {
	RegionCode               string     `json:"region_code"                bun:"regioncode"`
	RegionName               string     `json:"region_name"                bun:"regionname"`
	TotalDistricts           int64      `json:"total_districts"            bun:"total_districts"`
	TotalMeters              int64      `json:"total_meters"               bun:"total_meters"`
	MetersWithEvents         int64      `json:"meters_with_events"         bun:"meters_with_events"`
	MetersWithCoordinates    int64      `json:"meters_with_coordinates"    bun:"meters_with_coordinates"`
	MetersWithoutCoordinates int64      `json:"meters_without_coordinates" bun:"meters_without_coordinates"`
	TotalEvents              int64      `json:"total_events"               bun:"total_events"`
	LatestEventTime          *time.Time `json:"latest_event_time"          bun:"latest_event_time"`
}

// ── Queries ───────────────────────────────────────────────────────────────────

// GetSummary reads entirely from district_event_summary — never touches
// CustomerRecords. All values are pre-aggregated by post-load.
func (r *DashboardRepository) GetSummary(ctx context.Context) (*DashboardSummary, error) {
	var s DashboardSummary
	err := r.db.NewRaw(`
		SELECT
			SUM(total_meters)                AS total_meters,
			SUM(meters_with_coordinates)     AS meters_with_coordinates,
			SUM(meters_without_coordinates)  AS meters_without_coordinates,
			ROUND(
				SUM(meters_with_coordinates) * 100.0
				/ NULLIF(SUM(total_meters), 0), 2
			)                                AS coordinate_coverage_pct,
			SUM(meters_with_events)          AS meters_with_events,
			SUM(total_events)                AS total_events,
			COUNT(DISTINCT districtcode)     AS total_districts,
			COUNT(DISTINCT regioncode)       AS total_regions,
			SUM(prepaid_meters)              AS prepaid_meters,
			SUM(postpaid_meters)             AS postpaid_meters,
			SUM(active_contracts)            AS active_contracts,
			MAX(latest_event_time)           AS latest_event_time,
			MAX(updated_at)                  AS last_refreshed
		FROM district_event_summary
	`).Scan(ctx, &s)
	if err != nil {
		return nil, fmt.Errorf("dashboard summary: %w", err)
	}
	return &s, nil
}

func (r *DashboardRepository) GetDistricts(ctx context.Context, regionCode string) ([]DistrictSummary, error) {
	var districts []DistrictSummary
	q := r.db.NewRaw(`
		SELECT
			districtcode,
			districtname,
			regioncode,
			regionname,
			total_meters,
			meters_with_events,
			meters_with_coordinates,
			meters_without_coordinates,
			total_events,
			latest_event_time
		FROM district_event_summary
		WHERE (? = '' OR regioncode = ?)
		ORDER BY total_events DESC, districtname ASC
	`, regionCode, regionCode)

	if err := q.Scan(ctx, &districts); err != nil {
		return nil, fmt.Errorf("districts: %w", err)
	}
	return districts, nil
}

func (r *DashboardRepository) GetRegions(ctx context.Context) ([]RegionSummary, error) {
	var regions []RegionSummary
	err := r.db.NewRaw(`
		SELECT
			regioncode,
			regionname,
			COUNT(DISTINCT districtcode)     AS total_districts,
			SUM(total_meters)                AS total_meters,
			SUM(meters_with_events)          AS meters_with_events,
			SUM(meters_with_coordinates)     AS meters_with_coordinates,
			SUM(meters_without_coordinates)  AS meters_without_coordinates,
			SUM(total_events)                AS total_events,
			MAX(latest_event_time)           AS latest_event_time
		FROM district_event_summary
		GROUP BY regioncode, regionname
		ORDER BY total_events DESC
	`).Scan(ctx, &regions)
	if err != nil {
		return nil, fmt.Errorf("regions: %w", err)
	}
	return regions, nil
}
