package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"

	"github.com/kwaabs/m-events/internal/models"
)

// CustomerFilter holds optional query parameters for listing customers.
type CustomerFilter struct {
	Search             string
	RegionCode         string
	RegionName         string
	DistrictCode       string
	DistrictName       string
	ServiceType        string
	ServiceClass       string
	ServicePointNumber string
	CustomerType       string
	ContractStatus     string // contractstatus (Active/Inactive)
	CMSContractStatus  string // cmscontractstatus (Active Contract, PendingReplacement …)
	MeterMake          string
	MeterModel         string
	AccountType        string
	HasTamper          bool
	// Coordinate filters — useful for map gap analysis
	HasCoordinates bool // only customers where a plottable coordinate exists
	NoCoordinates  bool // only customers with no coordinate anywhere (cr or latest event)
	Page           int
	PageSize       int
}

// EventFilter holds pagination and optional filters for tamper event queries.
type EventFilter struct {
	// Pagination
	Page     int
	PageSize int

	// Filter by event_desc substring (case-insensitive)
	EventDesc string

	// Filter by event_code exact match
	EventCode string

	// Date range — both are optional and can be used independently
	From *time.Time // event_time >= From
	To   *time.Time // event_time <= To
}

type CustomerRepository struct {
	db                  *bun.DB
	coordinateColExists bool // set at startup by checkCoordinateColumn
}

func NewCustomerRepository(db *bun.DB) *CustomerRepository {
	r := &CustomerRepository{db: db}
	r.coordinateColExists = r.checkCoordinateColumn()
	return r
}

// checkCoordinateColumn returns true if migration 002 has been applied.
func (r *CustomerRepository) checkCoordinateColumn() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	_ = r.db.NewRaw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name  = 'CustomerRecords'
			  AND column_name = 'has_any_coordinates'
		)
	`).Scan(ctx, &exists)
	return exists
}

const latestEventLateral = `
	LATERAL (
		SELECT *
		FROM   "MMS_METER_TAMPER_EVENTS" mmte
		WHERE  mmte.meter_number = cr.meternumber
		ORDER  BY mmte.event_time DESC
		LIMIT  1
	) latest_event`

// ListWithLatestEvent returns a paginated list of customers each decorated
// with their most recent tamper event.
func (r *CustomerRepository) ListWithLatestEvent(
	ctx context.Context,
	f CustomerFilter,
) ([]models.CustomerRecord, int, error) {

	f = sanitizeFilter(f)

	// Guard: coordinate filters require migration 002 to have been applied.
	if (f.HasCoordinates || f.NoCoordinates) && !r.coordinateColExists {
		return nil, 0, fmt.Errorf(
			"coordinate filters are unavailable — run migrations/002_coordinate_flags.sql first",
		)
	}

	lateralJoin := "LEFT JOIN " + latestEventLateral + " ON TRUE"

	// ── Count query — no lateral join needed ─────────────────────────────────
	// Coordinate filters now use the has_any_coordinates column (plain index
	// scan) so the count never needs to touch the events table.
	countQ := r.db.NewSelect().
		TableExpr(`"CustomerRecords" AS cr`).
		ColumnExpr("COUNT(*)")
	applyFilters(countQ, f)

	total := 0
	if err := countQ.Scan(ctx, &total); err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	// ── Data query ───────────────────────────────────────────────────────────
	dataQ := r.db.NewSelect().
		TableExpr(`"CustomerRecords" AS cr`).
		ColumnExpr("cr.*").
		ColumnExpr(`
			latest_event.period        AS tamper_period,
			latest_event.meter_number  AS tamper_meter_number,
			latest_event.customer_name AS tamper_customer_name,
			latest_event.event_code    AS tamper_event_code,
			latest_event.event_desc    AS tamper_event_desc,
			latest_event.event_time    AS tamper_event_time,
			latest_event.latitude      AS tamper_latitude,
			latest_event.longitude     AS tamper_longitude,
			latest_event.counting      AS tamper_counting
		`).
		Join(lateralJoin)

	applyFilters(dataQ, f)

	offset := (f.Page - 1) * f.PageSize
	dataQ = dataQ.
		OrderExpr("cr.fullname ASC").
		Limit(f.PageSize).
		Offset(offset)

	rows, err := dataQ.Rows(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("data query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, 0, err
	}

	var customers []models.CustomerRecord
	for rows.Next() {
		cr, event, err := scanCustomerWithEvent(rows, cols)
		if err != nil {
			return nil, 0, err
		}
		cr.LatestTamperEvent = event
		customers = append(customers, cr)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// GetByAccountNumber returns a single customer with their latest tamper event.
func (r *CustomerRepository) GetByAccountNumber(
	ctx context.Context,
	accountNumber string,
) (*models.CustomerRecord, error) {

	rows, err := r.db.NewSelect().
		TableExpr(`"CustomerRecords" AS cr`).
		ColumnExpr("cr.*").
		ColumnExpr(`
			latest_event.period        AS tamper_period,
			latest_event.meter_number  AS tamper_meter_number,
			latest_event.customer_name AS tamper_customer_name,
			latest_event.event_code    AS tamper_event_code,
			latest_event.event_desc    AS tamper_event_desc,
			latest_event.event_time    AS tamper_event_time,
			latest_event.latitude      AS tamper_latitude,
			latest_event.longitude     AS tamper_longitude,
			latest_event.counting      AS tamper_counting
		`).
		Join("LEFT JOIN "+latestEventLateral+" ON TRUE").
		Where("cr.accountnumber = ?", accountNumber).
		Rows(ctx)

	if err != nil {
		return nil, fmt.Errorf("get by account: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		cr, event, err := scanCustomerWithEvent(rows, cols)
		if err != nil {
			return nil, err
		}
		cr.LatestTamperEvent = event
		return &cr, nil
	}

	return nil, nil
}

// GetByMeterNumber returns a single customer matched on meternumber.
func (r *CustomerRepository) GetByMeterNumber(
	ctx context.Context,
	meterNumber string,
) (*models.CustomerRecord, error) {

	rows, err := r.db.NewSelect().
		TableExpr(`"CustomerRecords" AS cr`).
		ColumnExpr("cr.*").
		ColumnExpr(`
			latest_event.period        AS tamper_period,
			latest_event.meter_number  AS tamper_meter_number,
			latest_event.customer_name AS tamper_customer_name,
			latest_event.event_code    AS tamper_event_code,
			latest_event.event_desc    AS tamper_event_desc,
			latest_event.event_time    AS tamper_event_time,
			latest_event.latitude      AS tamper_latitude,
			latest_event.longitude     AS tamper_longitude,
			latest_event.counting      AS tamper_counting
		`).
		Join("LEFT JOIN "+latestEventLateral+" ON TRUE").
		Where("cr.meternumber = ?", meterNumber).
		Rows(ctx)

	if err != nil {
		return nil, fmt.Errorf("get by meter: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		cr, event, err := scanCustomerWithEvent(rows, cols)
		if err != nil {
			return nil, err
		}
		cr.LatestTamperEvent = event
		return &cr, nil
	}

	return nil, nil
}

// TamperEventsByMeter returns paginated, filtered tamper events for a meter,
// newest first.
func (r *CustomerRepository) TamperEventsByMeter(
	ctx context.Context,
	meterNumber string,
	f EventFilter,
) ([]models.TamperEvent, int, error) {
	f = sanitizeEventFilter(f)

	countQ := r.db.NewSelect().
		TableExpr(`"MMS_METER_TAMPER_EVENTS"`).
		ColumnExpr("COUNT(*)").
		Where("meter_number = ?", meterNumber)
	applyEventFilters(countQ, f)

	total := 0
	if err := countQ.Scan(ctx, &total); err != nil {
		return nil, 0, fmt.Errorf("count tamper events by meter: %w", err)
	}

	dataQ := r.db.NewSelect().
		Model((*models.TamperEvent)(nil)).
		Where("meter_number = ?", meterNumber)
	applyEventFilters(dataQ, f)
	dataQ = dataQ.
		OrderExpr("event_time DESC").
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize)

	var events []models.TamperEvent
	if err := dataQ.Scan(ctx, &events); err != nil {
		return nil, 0, fmt.Errorf("tamper events by meter: %w", err)
	}

	return events, total, nil
}

// TamperEventsByAccount resolves the meter number from the account number then
// delegates to TamperEventsByMeter.
func (r *CustomerRepository) TamperEventsByAccount(
	ctx context.Context,
	accountNumber string,
	f EventFilter,
) ([]models.TamperEvent, int, string, error) {
	f = sanitizeEventFilter(f)

	var meterNumber string
	err := r.db.NewSelect().
		TableExpr(`"CustomerRecords"`).
		ColumnExpr("meternumber").
		Where("accountnumber = ?", accountNumber).
		Scan(ctx, &meterNumber)
	if err != nil {
		return nil, 0, "", fmt.Errorf("resolve meter for account: %w", err)
	}
	if meterNumber == "" {
		return nil, 0, "", nil
	}

	events, total, err := r.TamperEventsByMeter(ctx, meterNumber, f)
	return events, total, meterNumber, err
}

// ─── helpers ────────────────────────────────────────────────────────────────

func sanitizeFilter(f CustomerFilter) CustomerFilter {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 50
	}
	if f.PageSize > 200 {
		f.PageSize = 200
	}
	f.Search = strings.TrimSpace(f.Search)
	return f
}

func sanitizeEventFilter(f EventFilter) EventFilter {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 50
	}
	if f.PageSize > 500 {
		f.PageSize = 500
	}
	f.EventDesc = strings.TrimSpace(f.EventDesc)
	f.EventCode = strings.TrimSpace(f.EventCode)
	return f
}

func applyFilters(q *bun.SelectQuery, f CustomerFilter) {
	if f.Search != "" {
		like := "%" + strings.ToUpper(f.Search) + "%"
		q.Where(
			`(UPPER(cr.fullname) LIKE ? OR cr.accountnumber LIKE ? OR cr.meternumber LIKE ?)`,
			like, like, like,
		)
	}
	if f.RegionCode != "" {
		q.Where("cr.regioncode = ?", f.RegionCode)
	}
	if f.RegionName != "" {
		q.Where("UPPER(cr.regionname) LIKE ?", "%"+strings.ToUpper(f.RegionName)+"%")
	}
	if f.DistrictCode != "" {
		q.Where("cr.districtcode = ?", f.DistrictCode)
	}
	if f.DistrictName != "" {
		q.Where("UPPER(cr.districtname) LIKE ?", "%"+strings.ToUpper(f.DistrictName)+"%")
	}
	if f.ServiceType != "" {
		q.Where("cr.servicetype = ?", f.ServiceType)
	}
	if f.ServiceClass != "" {
		q.Where("cr.serviceclass = ?", f.ServiceClass)
	}
	if f.ServicePointNumber != "" {
		q.Where("cr.servicepointnumber = ?", f.ServicePointNumber)
	}
	if f.CustomerType != "" {
		q.Where("cr.customertype = ?", f.CustomerType)
	}
	if f.ContractStatus != "" {
		q.Where("cr.contractstatus = ?", f.ContractStatus)
	}
	if f.CMSContractStatus != "" {
		q.Where("cr.cmscontractstatus = ?", f.CMSContractStatus)
	}
	if f.MeterMake != "" {
		q.Where("UPPER(cr.metermake) LIKE ?", "%"+strings.ToUpper(f.MeterMake)+"%")
	}
	if f.MeterModel != "" {
		q.Where("UPPER(cr.metermodel) LIKE ?", "%"+strings.ToUpper(f.MeterModel)+"%")
	}
	if f.AccountType != "" {
		q.Where("cr.accounttype = ?", f.AccountType)
	}
	if f.HasTamper {
		// Use meter_summary (indexed PK lookup) instead of EXISTS on 29M events
		q.Where(`EXISTS (
			SELECT 1 FROM meter_summary ms
			WHERE ms.meternumber = cr.meternumber
		)`)
	}
	// Note: coordinate filters are only applied when the column exists.
	// If migration 002 has not been run the caller receives an error before
	// reaching here — see ListWithLatestEvent.
	if f.HasCoordinates {
		q.Where("cr.has_any_coordinates = TRUE")
	}
	if f.NoCoordinates {
		q.Where("cr.has_any_coordinates = FALSE")
	}
}

// applyEventFilters adds the optional event-level filters to any select query
// targeting MMS_METER_TAMPER_EVENTS.
func applyEventFilters(q *bun.SelectQuery, f EventFilter) {
	if f.EventDesc != "" {
		q.Where("UPPER(event_desc) LIKE ?", "%"+strings.ToUpper(f.EventDesc)+"%")
	}
	if f.EventCode != "" {
		q.Where("event_code = ?", f.EventCode)
	}
	if f.From != nil {
		q.Where("event_time >= ?", f.From)
	}
	if f.To != nil {
		q.Where("event_time <= ?", f.To)
	}
}

// scanCustomerWithEvent reads one row from a lateral-join result set.
func scanCustomerWithEvent(
	rows interface {
		Scan(...interface{}) error
	},
	cols []string,
) (models.CustomerRecord, *models.TamperEvent, error) {

	pointers := make([]interface{}, len(cols))

	var (
		tamperPeriod       *string
		tamperMeterNumber  *string
		tamperCustomerName *string
		tamperEventCode    *string
		tamperEventDesc    *string
		tamperEventTime    *string
		tamperLatitude     *float64
		tamperLongitude    *float64
		tamperCounting     *int64
	)

	crMap := make(map[string]*string)

	for i, col := range cols {
		switch col {
		case "tamper_period":
			pointers[i] = &tamperPeriod
		case "tamper_meter_number":
			pointers[i] = &tamperMeterNumber
		case "tamper_customer_name":
			pointers[i] = &tamperCustomerName
		case "tamper_event_code":
			pointers[i] = &tamperEventCode
		case "tamper_event_desc":
			pointers[i] = &tamperEventDesc
		case "tamper_event_time":
			pointers[i] = &tamperEventTime
		case "tamper_latitude":
			pointers[i] = &tamperLatitude
		case "tamper_longitude":
			pointers[i] = &tamperLongitude
		case "tamper_counting":
			pointers[i] = &tamperCounting
		default:
			var v interface{}
			crMap[col] = nil
			pointers[i] = &v
			_ = crMap
			pointers[i] = new(interface{})
		}
	}

	if err := rows.Scan(pointers...); err != nil {
		return models.CustomerRecord{}, nil, fmt.Errorf("scan row: %w", err)
	}

	cr := models.CustomerRecord{}
	crValues := map[string]interface{}{}
	for i, col := range cols {
		if !strings.HasPrefix(col, "tamper_") {
			if iface, ok := pointers[i].(*interface{}); ok {
				crValues[col] = *iface
			}
		}
	}
	mapToCR(crValues, &cr)

	var event *models.TamperEvent
	if tamperMeterNumber != nil {
		event = &models.TamperEvent{}
		event.MeterNumber = *tamperMeterNumber
		if tamperPeriod != nil {
			event.Period = *tamperPeriod
		}
		if tamperCustomerName != nil {
			event.CustomerName = *tamperCustomerName
		}
		if tamperEventCode != nil {
			event.EventCode = *tamperEventCode
		}
		if tamperEventDesc != nil {
			event.EventDesc = *tamperEventDesc
		}
		if tamperEventTime != nil {
			t, err := parseTime(*tamperEventTime)
			if err == nil {
				event.EventTime = t
			}
		}
		event.Latitude = tamperLatitude
		event.Longitude = tamperLongitude
		event.Counting = tamperCounting
	}

	return cr, event, nil
}
