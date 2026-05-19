//package models
//
//import (
//	"time"
//
//	"github.com/uptrace/bun"
//)
//
//// CustomerRecord maps to the "CustomerRecords" table.
//// Column names match the exact lowercase column names in Postgres.
//// bun:"table" sets the quoted table name.
//type CustomerRecord struct {
//	bun.BaseModel `bun:"table:CustomerRecords,alias:cr"`
//
//	ECGKey               string     `bun:"ecgkey,pk" json:"ecgkey"`
//	ID                   string     `bun:"_id" json:"_id"`
//	RegionCode           string     `bun:"regioncode" json:"regioncode"`
//	RegionName           string     `bun:"regionname" json:"regionname"`
//	DistrictCode         string     `bun:"districtcode" json:"districtcode"`
//	DistrictName         string     `bun:"districtname" json:"districtname"`
//	ServiceType          string     `bun:"servicetype" json:"servicetype"`
//	ServiceClass         string     `bun:"serviceclass" json:"serviceclass"`
//	TariffClassCode      string     `bun:"tariffclasscode" json:"tariffclasscode"`
//	TariffClassName      string     `bun:"tariffclassname" json:"tariffclassname"`
//	Community            string     `bun:"community" json:"community"`
//	ContractedDemand     *float64   `bun:"contracteddemand" json:"contracteddemand"`
//	QRCode               string     `bun:"qrcode" json:"qrcode"`
//	CreatedAt            time.Time  `bun:"createdat" json:"createdat"`
//	UpdatedAt            time.Time  `bun:"updatedat" json:"updatedat"`
//	Address              string     `bun:"address" json:"address"`
//	GhanaPostAddress     string     `bun:"ghanapostaddress" json:"ghanapostaddress"`
//	StreetName           string     `bun:"streetname" json:"streetname"`
//	HouseNumber          string     `bun:"housenumber" json:"housenumber"`
//	GeoLocationType      *string    `bun:"geolocationtype" json:"geolocationtype"`
//	GeoLocationID        *string    `bun:"geolocationid" json:"geolocationid"`
//	Latitude             *float64   `bun:"latitude" json:"latitude"`
//	Longitude            *float64   `bun:"longitude" json:"longitude"`
//	EmailAddress         string     `bun:"emailaddress" json:"emailaddress"`
//	PhoneNumber          string     `bun:"phonenumber" json:"phonenumber"`
//	ProfileImageURL      string     `bun:"profileimageurl" json:"profileimageurl"`
//	FullName             string     `bun:"fullname" json:"fullname"`
//	GhanaCardNumber      *string    `bun:"ghanacardnumber" json:"ghanacardnumber"`
//	PayerFullName        string     `bun:"payerfullname" json:"payerfullname"`
//	PayerPrimaryPhone    string     `bun:"payerprimaryphonenumber" json:"payerprimaryphonenumber"`
//	PayerFirstPhone      *string    `bun:"payerfirstphone" json:"payerfirstphone"`
//	PayerSecondPhone     *string    `bun:"payersecondphone" json:"payersecondphone"`
//	ServicePointNumber   string     `bun:"servicepointnumber" json:"servicepointnumber"`
//	AccountNumber        string     `bun:"accountnumber" json:"accountnumber"`
//	ContractStatus       string     `bun:"contractstatus" json:"contractstatus"`
//	IsMigratedData       bool       `bun:"ismigrateddata" json:"ismigrateddata"`
//	MeterType            string     `bun:"metertype" json:"metertype"`
//	MeterLocation        string     `bun:"meterlocation" json:"meterlocation"`
//	Activity             *string    `bun:"activity" json:"activity"`
//	SubActivity          *string    `bun:"subactivity" json:"subactivity"`
//	CustomerType         string     `bun:"customertype" json:"customertype"`
//	IsRegularized        bool       `bun:"isregularized" json:"isregularized"`
//	LastReadingValue     *int64     `bun:"lastreadingvalue" json:"lastreadingvalue"`
//	GeoLocationCoords    string     `bun:"geolocationcoordinates" json:"geolocationcoordinates"`
//	PayerPhoneNumbers    string     `bun:"payerphonenumbers" json:"payerphonenumbers"`
//	BlockCode            string     `bun:"blockcode" json:"blockcode"`
//	BlockName            string     `bun:"blockname" json:"blockname"`
//	RoundCode            string     `bun:"roundcode" json:"roundcode"`
//	RoundName            string     `bun:"roundname" json:"roundname"`
//	CMSContractStatus    string     `bun:"cmscontractstatus" json:"cmscontractstatus"`
//	Code                 string     `bun:"code" json:"code"`
//	GeoCode              string     `bun:"geocode" json:"geocode"`
//	MeterMake            string     `bun:"metermake" json:"metermake"`
//	MeterModel           string     `bun:"metermodel" json:"metermodel"`
//	MeterNumber          string     `bun:"meternumber" json:"meternumber"`
//	MeterPhase           string     `bun:"meterphase" json:"meterphase"`
//	PropertyCode         string     `bun:"propertycode" json:"propertycode"`
//	PlotCode             string     `bun:"plotcode" json:"plotcode"`
//	Ministry             *string    `bun:"ministry" json:"ministry"`
//	MDA                  string     `bun:"mda" json:"mda"`
//	LastReadingDate      *time.Time `bun:"lastreadingdate" json:"lastreadingdate"`
//	InstallationDate     *time.Time `bun:"installationdate" json:"installationdate"`
//	LastBillAmount       *float64   `bun:"lastbillamount" json:"lastbillamount"`
//	LastBillConsumption  *float64   `bun:"lastbillconsumption" json:"lastbillconsumption"`
//	LastPaymentDate      *string    `bun:"lastpaymentdate" json:"lastpaymentdate"`
//	LastPaymentAmount    *float64   `bun:"lastpaymentamount" json:"lastpaymentamount"`
//	CurrentBalance       *float64   `bun:"currentbalance" json:"currentbalance"`
//	Nationality          string     `bun:"nationality" json:"nationality"`
//	PropertyOwnerName    string     `bun:"propertyownerfullname" json:"propertyownerfullname"`
//	PropertyOwnerPhone   string     `bun:"propertyownerphonenumber" json:"propertyownerphonenumber"`
//	AccountType          string     `bun:"accounttype" json:"accounttype"`
//	IsAMR                bool       `bun:"isamr" json:"isamr"`
//	Offset               string     `bun:"offset" json:"offset"`
//	IDNumber             *string    `bun:"idnumber" json:"idnumber"`
//	MinistryCode         *string    `bun:"ministrycode" json:"ministrycode"`
//	MinistryName         *string    `bun:"ministryname" json:"ministryname"`
//	MDACode              *string    `bun:"mdacode" json:"mdacode"`
//	MDAName              *string    `bun:"mdaname" json:"mdaname"`
//	ExtraAttributesName  *string    `bun:"extraattributesname" json:"extraattributesname"`
//	ExtraAttributesValue bool       `bun:"extraattributesvalue" json:"extraattributesvalue"`
//	ExtraAttributesID    *string    `bun:"extraattributesid" json:"extraattributesid"`
//	LastBillDate         *time.Time `bun:"lastbilldate" json:"lastbilldate"`
//
//	// Coordinate flags — maintained by DB triggers, never null
//	HasOwnCoordinates   bool `bun:"has_own_coordinates"   json:"has_own_coordinates"`
//	HasEventCoordinates bool `bun:"has_event_coordinates" json:"has_event_coordinates"`
//	HasAnyCoordinates   bool `bun:"has_any_coordinates"   json:"has_any_coordinates"`
//
//	// Populated via JOIN / sub-query — not a DB column
//	LatestTamperEvent *TamperEvent `bun:"-" json:"latest_tamper_event,omitempty"`
//}
//
//// TamperEvent maps to the "MMS_METER_TAMPER_EVENTS" table.
//type TamperEvent struct {
//	bun.BaseModel `bun:"table:MMS_METER_TAMPER_EVENTS,alias:mmte"`
//
//	Period       string    `bun:"period" json:"period"`
//	MeterNumber  string    `bun:"meter_number,pk" json:"meter_number"`
//	CustomerName string    `bun:"customer_name" json:"customer_name"`
//	EventCode    string    `bun:"event_code" json:"event_code"`
//	EventDesc    string    `bun:"event_desc" json:"event_desc"`
//	EventTime    time.Time `bun:"event_time" json:"event_time"`
//	Latitude     *float64  `bun:"latitude" json:"latitude"`
//	Longitude    *float64  `bun:"longitude" json:"longitude"`
//	Counting     *int64    `bun:"counting" json:"counting"`
//}
//
//// ─── Response / DTO types ────────────────────────────────────────────────────
//
//// CustomerWithEvent is the shape returned to the MapLibre frontend.
//// It flattens what the map needs (coordinates, identity, latest event).
//type CustomerWithEvent struct {
//	// Identity
//	ECGKey             string `json:"ecgkey"`
//	AccountNumber      string `json:"account_number"`
//	ServicePointNumber string `json:"service_point_number"`
//	FullName           string `json:"full_name"`
//	PhoneNumber        string `json:"phone_number"`
//
//	// Service
//	ServiceType    string `json:"service_type"`
//	ServiceClass   string `json:"service_class"`
//	TariffClass    string `json:"tariff_class"`
//	ContractStatus string `json:"contract_status"`
//	MeterNumber    string `json:"meter_number"`
//	MeterType      string `json:"meter_type"`
//	MeterMake      string `json:"meter_make"`
//	MeterModel     string `json:"meter_model"`
//
//	// Location
//	Address           string   `json:"address"`
//	StreetName        string   `json:"street_name"`
//	HouseNumber       string   `json:"house_number"`
//	RegionName        string   `json:"region_name"`
//	DistrictName      string   `json:"district_name"`
//	Latitude          *float64 `json:"latitude"`
//	Longitude         *float64 `json:"longitude"`
//	HasAnyCoordinates bool     `json:"has_any_coordinates"`
//
//	// Billing snapshot
//	LastBillAmount    *float64 `json:"last_bill_amount"`
//	CurrentBalance    *float64 `json:"current_balance"`
//	LastPaymentDate   *string  `json:"last_payment_date"`
//	LastPaymentAmount *float64 `json:"last_payment_amount"`
//
//	// Latest tamper event (nil if none)
//	LatestTamperEvent *TamperEventSummary `json:"latest_tamper_event"`
//}
//
//// TamperEventSummary is a lighter view used inside CustomerWithEvent.
//type TamperEventSummary struct {
//	EventCode string    `json:"event_code"`
//	EventDesc string    `json:"event_desc"`
//	EventTime time.Time `json:"event_time"`
//	Period    string    `json:"period"`
//	Latitude  *float64  `json:"latitude"`
//	Longitude *float64  `json:"longitude"`
//}
//
//// ToCustomerWithEvent maps a CustomerRecord (with optional LatestTamperEvent)
//// into the API response DTO.
//func ToCustomerWithEvent(cr CustomerRecord) CustomerWithEvent {
//	c := CustomerWithEvent{
//		ECGKey:             cr.ECGKey,
//		AccountNumber:      cr.AccountNumber,
//		ServicePointNumber: cr.ServicePointNumber,
//		FullName:           cr.FullName,
//		PhoneNumber:        cr.PhoneNumber,
//		ServiceType:        cr.ServiceType,
//		ServiceClass:       cr.ServiceClass,
//		TariffClass:        cr.TariffClassName,
//		ContractStatus:     cr.ContractStatus,
//		MeterNumber:        cr.MeterNumber,
//		MeterType:          cr.MeterType,
//		MeterMake:          cr.MeterMake,
//		MeterModel:         cr.MeterModel,
//		Address:            cr.Address,
//		StreetName:         cr.StreetName,
//		HouseNumber:        cr.HouseNumber,
//		RegionName:         cr.RegionName,
//		DistrictName:       cr.DistrictName,
//		Latitude:           cr.Latitude,
//		Longitude:          cr.Longitude,
//		HasAnyCoordinates:  cr.HasAnyCoordinates,
//		LastBillAmount:     cr.LastBillAmount,
//		CurrentBalance:     cr.CurrentBalance,
//		LastPaymentDate:    cr.LastPaymentDate,
//		LastPaymentAmount:  cr.LastPaymentAmount,
//	}
//
//	if cr.LatestTamperEvent != nil {
//		e := cr.LatestTamperEvent
//		// Prefer event coordinates when customer record has none
//		lat, lon := cr.Latitude, cr.Longitude
//		if lat == nil {
//			lat = e.Latitude
//		}
//		if lon == nil {
//			lon = e.Longitude
//		}
//		c.Latitude = lat
//		c.Longitude = lon
//
//		c.LatestTamperEvent = &TamperEventSummary{
//			EventCode: e.EventCode,
//			EventDesc: e.EventDesc,
//			EventTime: e.EventTime,
//			Period:    e.Period,
//			Latitude:  e.Latitude,
//			Longitude: e.Longitude,
//		}
//	}
//
//	return c
//}
//
//// PaginatedResponse wraps a list result with pagination metadata.
//type PaginatedResponse[T any] struct {
//	Data       []T `json:"data"`
//	Total      int `json:"total"`
//	Page       int `json:"page"`
//	PageSize   int `json:"page_size"`
//	TotalPages int `json:"total_pages"`
//}

package models

import (
	"time"

	"github.com/uptrace/bun"
)

// CustomerRecord maps to the "CustomerRecords" table.
// Column names match the exact lowercase column names in Postgres.
// bun:"table" sets the quoted table name.
type CustomerRecord struct {
	bun.BaseModel `bun:"table:CustomerRecords,alias:cr"`

	ECGKey               string     `bun:"ecgkey,pk" json:"ecgkey"`
	ID                   string     `bun:"_id" json:"_id"`
	RegionCode           string     `bun:"regioncode" json:"regioncode"`
	RegionName           string     `bun:"regionname" json:"regionname"`
	DistrictCode         string     `bun:"districtcode" json:"districtcode"`
	DistrictName         string     `bun:"districtname" json:"districtname"`
	ServiceType          string     `bun:"servicetype" json:"servicetype"`
	ServiceClass         string     `bun:"serviceclass" json:"serviceclass"`
	TariffClassCode      string     `bun:"tariffclasscode" json:"tariffclasscode"`
	TariffClassName      string     `bun:"tariffclassname" json:"tariffclassname"`
	Community            string     `bun:"community" json:"community"`
	ContractedDemand     *float64   `bun:"contracteddemand" json:"contracteddemand"`
	QRCode               string     `bun:"qrcode" json:"qrcode"`
	CreatedAt            time.Time  `bun:"createdat" json:"createdat"`
	UpdatedAt            time.Time  `bun:"updatedat" json:"updatedat"`
	Address              string     `bun:"address" json:"address"`
	GhanaPostAddress     string     `bun:"ghanapostaddress" json:"ghanapostaddress"`
	StreetName           string     `bun:"streetname" json:"streetname"`
	HouseNumber          string     `bun:"housenumber" json:"housenumber"`
	GeoLocationType      *string    `bun:"geolocationtype" json:"geolocationtype"`
	GeoLocationID        *string    `bun:"geolocationid" json:"geolocationid"`
	Latitude             *float64   `bun:"latitude" json:"latitude"`
	Longitude            *float64   `bun:"longitude" json:"longitude"`
	EmailAddress         string     `bun:"emailaddress" json:"emailaddress"`
	PhoneNumber          string     `bun:"phonenumber" json:"phonenumber"`
	ProfileImageURL      string     `bun:"profileimageurl" json:"profileimageurl"`
	FullName             string     `bun:"fullname" json:"fullname"`
	GhanaCardNumber      *string    `bun:"ghanacardnumber" json:"ghanacardnumber"`
	PayerFullName        string     `bun:"payerfullname" json:"payerfullname"`
	PayerPrimaryPhone    string     `bun:"payerprimaryphonenumber" json:"payerprimaryphonenumber"`
	PayerFirstPhone      *string    `bun:"payerfirstphone" json:"payerfirstphone"`
	PayerSecondPhone     *string    `bun:"payersecondphone" json:"payersecondphone"`
	ServicePointNumber   string     `bun:"servicepointnumber" json:"servicepointnumber"`
	AccountNumber        string     `bun:"accountnumber" json:"accountnumber"`
	ContractStatus       string     `bun:"contractstatus" json:"contractstatus"`
	IsMigratedData       bool       `bun:"ismigrateddata" json:"ismigrateddata"`
	MeterType            string     `bun:"metertype" json:"metertype"`
	MeterLocation        string     `bun:"meterlocation" json:"meterlocation"`
	Activity             *string    `bun:"activity" json:"activity"`
	SubActivity          *string    `bun:"subactivity" json:"subactivity"`
	CustomerType         string     `bun:"customertype" json:"customertype"`
	IsRegularized        bool       `bun:"isregularized" json:"isregularized"`
	LastReadingValue     *int64     `bun:"lastreadingvalue" json:"lastreadingvalue"`
	GeoLocationCoords    string     `bun:"geolocationcoordinates" json:"geolocationcoordinates"`
	PayerPhoneNumbers    string     `bun:"payerphonenumbers" json:"payerphonenumbers"`
	BlockCode            string     `bun:"blockcode" json:"blockcode"`
	BlockName            string     `bun:"blockname" json:"blockname"`
	RoundCode            string     `bun:"roundcode" json:"roundcode"`
	RoundName            string     `bun:"roundname" json:"roundname"`
	CMSContractStatus    string     `bun:"cmscontractstatus" json:"cmscontractstatus"`
	Code                 string     `bun:"code" json:"code"`
	GeoCode              string     `bun:"geocode" json:"geocode"`
	MeterMake            string     `bun:"metermake" json:"metermake"`
	MeterModel           string     `bun:"metermodel" json:"metermodel"`
	MeterNumber          string     `bun:"meternumber" json:"meternumber"`
	MeterPhase           string     `bun:"meterphase" json:"meterphase"`
	PropertyCode         string     `bun:"propertycode" json:"propertycode"`
	PlotCode             string     `bun:"plotcode" json:"plotcode"`
	Ministry             *string    `bun:"ministry" json:"ministry"`
	MDA                  string     `bun:"mda" json:"mda"`
	LastReadingDate      *time.Time `bun:"lastreadingdate" json:"lastreadingdate"`
	InstallationDate     *time.Time `bun:"installationdate" json:"installationdate"`
	LastBillAmount       *float64   `bun:"lastbillamount" json:"lastbillamount"`
	LastBillConsumption  *float64   `bun:"lastbillconsumption" json:"lastbillconsumption"`
	LastPaymentDate      *string    `bun:"lastpaymentdate" json:"lastpaymentdate"`
	LastPaymentAmount    *float64   `bun:"lastpaymentamount" json:"lastpaymentamount"`
	CurrentBalance       *float64   `bun:"currentbalance" json:"currentbalance"`
	Nationality          string     `bun:"nationality" json:"nationality"`
	PropertyOwnerName    string     `bun:"propertyownerfullname" json:"propertyownerfullname"`
	PropertyOwnerPhone   string     `bun:"propertyownerphonenumber" json:"propertyownerphonenumber"`
	AccountType          string     `bun:"accounttype" json:"accounttype"`
	IsAMR                bool       `bun:"isamr" json:"isamr"`
	Offset               string     `bun:"offset" json:"offset"`
	IDNumber             *string    `bun:"idnumber" json:"idnumber"`
	MinistryCode         *string    `bun:"ministrycode" json:"ministrycode"`
	MinistryName         *string    `bun:"ministryname" json:"ministryname"`
	MDACode              *string    `bun:"mdacode" json:"mdacode"`
	MDAName              *string    `bun:"mdaname" json:"mdaname"`
	ExtraAttributesName  *string    `bun:"extraattributesname" json:"extraattributesname"`
	ExtraAttributesValue bool       `bun:"extraattributesvalue" json:"extraattributesvalue"`
	ExtraAttributesID    *string    `bun:"extraattributesid" json:"extraattributesid"`
	LastBillDate         *time.Time `bun:"lastbilldate" json:"lastbilldate"`

	// Coordinate flags — maintained by DB triggers, never null
	HasOwnCoordinates   bool `bun:"has_own_coordinates"   json:"has_own_coordinates"`
	HasEventCoordinates bool `bun:"has_event_coordinates" json:"has_event_coordinates"`
	HasAnyCoordinates   bool `bun:"has_any_coordinates"   json:"has_any_coordinates"`

	// Populated via JOIN / sub-query — not a DB column
	LatestTamperEvent *TamperEvent `bun:"-" json:"latest_tamper_event,omitempty"`
}

// TamperEvent maps to the "MMS_METER_TAMPER_EVENTS" table.
type TamperEvent struct {
	bun.BaseModel `bun:"table:MMS_METER_TAMPER_EVENTS,alias:mmte"`

	Period       string    `bun:"period" json:"period"`
	MeterNumber  string    `bun:"meter_number,pk" json:"meter_number"`
	CustomerName string    `bun:"customer_name" json:"customer_name"`
	EventCode    string    `bun:"event_code" json:"event_code"`
	EventDesc    string    `bun:"event_desc" json:"event_desc"`
	EventTime    time.Time `bun:"event_time" json:"event_time"`
	Latitude     *float64  `bun:"latitude" json:"latitude"`
	Longitude    *float64  `bun:"longitude" json:"longitude"`
	Counting     *int64    `bun:"counting" json:"counting"`
}

// ─── Response / DTO types ────────────────────────────────────────────────────

// CustomerWithEvent is the shape returned to the MapLibre frontend.
// It flattens what the map needs (coordinates, identity, latest event).
type CustomerWithEvent struct {
	// Identity
	ECGKey             string `json:"ecgkey"`
	AccountNumber      string `json:"account_number"`
	ServicePointNumber string `json:"service_point_number"`
	FullName           string `json:"full_name"`
	PhoneNumber        string `json:"phone_number"`

	// Service
	ServiceType    string `json:"service_type"`
	ServiceClass   string `json:"service_class"`
	TariffClass    string `json:"tariff_class"`
	ContractStatus string `json:"contract_status"`
	MeterNumber    string `json:"meter_number"`
	MeterType      string `json:"meter_type"`
	MeterMake      string `json:"meter_make"`
	MeterModel     string `json:"meter_model"`

	// Location
	Address           string   `json:"address"`
	StreetName        string   `json:"street_name"`
	HouseNumber       string   `json:"house_number"`
	RegionName        string   `json:"region_name"`
	DistrictName      string   `json:"district_name"`
	Latitude          *float64 `json:"latitude"`
	Longitude         *float64 `json:"longitude"`
	HasAnyCoordinates bool     `json:"has_any_coordinates"`

	// Billing snapshot
	LastBillAmount    *float64 `json:"last_bill_amount"`
	CurrentBalance    *float64 `json:"current_balance"`
	LastPaymentDate   *string  `json:"last_payment_date"`
	LastPaymentAmount *float64 `json:"last_payment_amount"`

	// Latest tamper event (nil if none)
	LatestTamperEvent *TamperEventSummary `json:"latest_tamper_event"`
}

// TamperEventSummary is a lighter view used inside CustomerWithEvent.
type TamperEventSummary struct {
	EventCode string    `json:"event_code"`
	EventDesc string    `json:"event_desc"`
	EventTime time.Time `json:"event_time"`
	Period    string    `json:"period"`
	Latitude  *float64  `json:"latitude"`
	Longitude *float64  `json:"longitude"`
}

// ToCustomerWithEvent maps a CustomerRecord (with optional LatestTamperEvent)
// into the API response DTO.
func ToCustomerWithEvent(cr CustomerRecord) CustomerWithEvent {
	c := CustomerWithEvent{
		ECGKey:             cr.ECGKey,
		AccountNumber:      cr.AccountNumber,
		ServicePointNumber: cr.ServicePointNumber,
		FullName:           cr.FullName,
		PhoneNumber:        cr.PhoneNumber,
		ServiceType:        cr.ServiceType,
		ServiceClass:       cr.ServiceClass,
		TariffClass:        cr.TariffClassName,
		ContractStatus:     cr.ContractStatus,
		MeterNumber:        cr.MeterNumber,
		MeterType:          cr.MeterType,
		MeterMake:          cr.MeterMake,
		MeterModel:         cr.MeterModel,
		Address:            cr.Address,
		StreetName:         cr.StreetName,
		HouseNumber:        cr.HouseNumber,
		RegionName:         cr.RegionName,
		DistrictName:       cr.DistrictName,
		// Source data has lat/lng swapped — correct in the API response
		Latitude:          cr.Longitude,
		Longitude:         cr.Latitude,
		HasAnyCoordinates: cr.HasAnyCoordinates,
		LastBillAmount:    cr.LastBillAmount,
		CurrentBalance:    cr.CurrentBalance,
		LastPaymentDate:   cr.LastPaymentDate,
		LastPaymentAmount: cr.LastPaymentAmount,
	}

	if cr.LatestTamperEvent != nil {
		e := cr.LatestTamperEvent
		// Source data has lat/lng swapped — correct in the API response
		// Use cr.Longitude as latitude, cr.Latitude as longitude
		lat, lon := cr.Longitude, cr.Latitude
		if lat == nil {
			lat = e.Longitude // event lat is also swapped
		}
		if lon == nil {
			lon = e.Latitude
		}
		c.Latitude = lat
		c.Longitude = lon

		c.LatestTamperEvent = &TamperEventSummary{
			EventCode: e.EventCode,
			EventDesc: e.EventDesc,
			EventTime: e.EventTime,
			Period:    e.Period,
			Latitude:  e.Longitude, // swap for correct output
			Longitude: e.Latitude,
		}
	}

	return c
}

// PaginatedResponse wraps a list result with pagination metadata.
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}
