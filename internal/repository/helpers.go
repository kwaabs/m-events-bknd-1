package repository

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kwaabs/m-events/internal/models"
)

// parseTime tries a set of common Postgres timestamp formats.
func parseTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}

// toString safely coerces an interface{} database value to *string.
func toString(v interface{}) *string {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return &t
	case []byte:
		s := string(t)
		if s == "" {
			return nil
		}
		return &s
	default:
		s := fmt.Sprintf("%v", v)
		return &s
	}
}

func toStringVal(v interface{}) string {
	if s := toString(v); s != nil {
		return *s
	}
	return ""
}

func toFloat64(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case float64:
		return &t
	case float32:
		f := float64(t)
		return &f
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return nil
		}
		return &f
	case []byte:
		f, err := strconv.ParseFloat(string(t), 64)
		if err != nil {
			return nil
		}
		return &f
	}
	return nil
}

func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1" || t == "t"
	case int64:
		return t != 0
	}
	return false
}

func toTimePtr(v interface{}) *time.Time {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case time.Time:
		return &t
	case string:
		parsed, err := parseTime(t)
		if err != nil {
			return nil
		}
		return &parsed
	case []byte:
		parsed, err := parseTime(string(t))
		if err != nil {
			return nil
		}
		return &parsed
	}
	return nil
}

func toInt64Ptr(v interface{}) *int64 {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case int64:
		return &t
	case int32:
		i := int64(t)
		return &i
	case float64:
		i := int64(t)
		return &i
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil
		}
		return &i
	}
	return nil
}

// mapToCR fills a CustomerRecord from the generic column-value map produced
// by scanCustomerWithEvent. This avoids a second DB round-trip while still
// giving us proper typing.
func mapToCR(m map[string]interface{}, cr *models.CustomerRecord) {
	cr.ECGKey = toStringVal(m["ecgkey"])
	cr.ID = toStringVal(m["_id"])
	cr.RegionCode = toStringVal(m["regioncode"])
	cr.RegionName = toStringVal(m["regionname"])
	cr.DistrictCode = toStringVal(m["districtcode"])
	cr.DistrictName = toStringVal(m["districtname"])
	cr.ServiceType = toStringVal(m["servicetype"])
	cr.ServiceClass = toStringVal(m["serviceclass"])
	cr.TariffClassCode = toStringVal(m["tariffclasscode"])
	cr.TariffClassName = toStringVal(m["tariffclassname"])
	cr.Community = toStringVal(m["community"])
	cr.ContractedDemand = toFloat64(m["contracteddemand"])
	cr.QRCode = toStringVal(m["qrcode"])
	cr.Address = toStringVal(m["address"])
	cr.GhanaPostAddress = toStringVal(m["ghanapostaddress"])
	cr.StreetName = toStringVal(m["streetname"])
	cr.HouseNumber = toStringVal(m["housenumber"])
	cr.GeoLocationType = toString(m["geolocationtype"])
	cr.GeoLocationID = toString(m["geolocationid"])
	cr.Latitude = toFloat64(m["latitude"])
	cr.Longitude = toFloat64(m["longitude"])
	cr.EmailAddress = toStringVal(m["emailaddress"])
	cr.PhoneNumber = toStringVal(m["phonenumber"])
	cr.ProfileImageURL = toStringVal(m["profileimageurl"])
	cr.FullName = toStringVal(m["fullname"])
	cr.GhanaCardNumber = toString(m["ghanacardnumber"])
	cr.PayerFullName = toStringVal(m["payerfullname"])
	cr.PayerPrimaryPhone = toStringVal(m["payerprimaryphonenumber"])
	cr.PayerFirstPhone = toString(m["payerfirstphone"])
	cr.PayerSecondPhone = toString(m["payersecondphone"])
	cr.ServicePointNumber = toStringVal(m["servicepointnumber"])
	cr.AccountNumber = toStringVal(m["accountnumber"])
	cr.ContractStatus = toStringVal(m["contractstatus"])
	cr.IsMigratedData = toBool(m["ismigrateddata"])
	cr.MeterType = toStringVal(m["metertype"])
	cr.MeterLocation = toStringVal(m["meterlocation"])
	cr.Activity = toString(m["activity"])
	cr.SubActivity = toString(m["subactivity"])
	cr.CustomerType = toStringVal(m["customertype"])
	cr.IsRegularized = toBool(m["isregularized"])
	cr.LastReadingValue = toInt64Ptr(m["lastreadingvalue"])
	cr.GeoLocationCoords = toStringVal(m["geolocationcoordinates"])
	cr.PayerPhoneNumbers = toStringVal(m["payerphonenumbers"])
	cr.BlockCode = toStringVal(m["blockcode"])
	cr.BlockName = toStringVal(m["blockname"])
	cr.RoundCode = toStringVal(m["roundcode"])
	cr.RoundName = toStringVal(m["roundname"])
	cr.CMSContractStatus = toStringVal(m["cmscontractstatus"])
	cr.Code = toStringVal(m["code"])
	cr.GeoCode = toStringVal(m["geocode"])
	cr.MeterMake = toStringVal(m["metermake"])
	cr.MeterModel = toStringVal(m["metermodel"])
	cr.MeterNumber = toStringVal(m["meternumber"])
	cr.MeterPhase = toStringVal(m["meterphase"])
	cr.PropertyCode = toStringVal(m["propertycode"])
	cr.PlotCode = toStringVal(m["plotcode"])
	cr.Ministry = toString(m["ministry"])
	cr.MDA = toStringVal(m["mda"])
	cr.LastReadingDate = toTimePtr(m["lastreadingdate"])
	cr.InstallationDate = toTimePtr(m["installationdate"])
	cr.LastBillAmount = toFloat64(m["lastbillamount"])
	cr.LastBillConsumption = toFloat64(m["lastbillconsumption"])
	cr.LastPaymentDate = toString(m["lastpaymentdate"])
	cr.LastPaymentAmount = toFloat64(m["lastpaymentamount"])
	cr.CurrentBalance = toFloat64(m["currentbalance"])
	cr.Nationality = toStringVal(m["nationality"])
	cr.PropertyOwnerName = toStringVal(m["propertyownerfullname"])
	cr.PropertyOwnerPhone = toStringVal(m["propertyownerphonenumber"])
	cr.AccountType = toStringVal(m["accounttype"])
	cr.IsAMR = toBool(m["isamr"])
	cr.Offset = toStringVal(m["offset"])
	cr.IDNumber = toString(m["idnumber"])
	cr.MinistryCode = toString(m["ministrycode"])
	cr.MinistryName = toString(m["ministryname"])
	cr.MDACode = toString(m["mdacode"])
	cr.MDAName = toString(m["mdaname"])
	cr.ExtraAttributesName = toString(m["extraattributesname"])
	cr.ExtraAttributesValue = toBool(m["extraattributesvalue"])
	cr.ExtraAttributesID = toString(m["extraattributesid"])
	cr.LastBillDate = toTimePtr(m["lastbilldate"])
	cr.HasOwnCoordinates = toBool(m["has_own_coordinates"])
	cr.HasEventCoordinates = toBool(m["has_event_coordinates"])
	cr.HasAnyCoordinates = toBool(m["has_any_coordinates"])
}
