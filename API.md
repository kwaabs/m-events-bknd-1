# ECG Tamper Map — API Reference

**Base URL:** `http://<host>:8080`  
**All `/api/v1/*` routes require a valid JWT** except `POST /api/v1/auth/login`.  
**Content-Type:** `application/json` on all requests and responses.

---

## Table of Contents

1. [Authentication](#authentication)
    - [Login](#post-apiv1authlogin)
    - [Inspect Token](#get-apiv1authme)
2. [Customers](#customers)
    - [List Customers](#get-apiv1customers)
    - [Get by Account Number](#get-apiv1customersaccountaccountnumber)
    - [Get by Meter Number](#get-apiv1customersmetermeternumber)
3. [Tamper Events](#tamper-events)
    - [Events by Meter Number](#get-apiv1customersmetermeternumberevents)
    - [Events by Account Number](#get-apiv1customersaccountaccountnumberevents)
4. [Health Check](#health-check)
5. [Filter Reference](#filter-reference)
6. [Response Shapes](#response-shapes)
7. [Error Responses](#error-responses)
8. [MapLibre Integration](#maplibre-integration)

---

## Authentication

The API uses **JWT Bearer tokens**. Tokens are issued by logging in with your
Active Directory credentials. Every protected request must include the token in
the `Authorization` header:

```
Authorization: Bearer <token>
```

Tokens expire after **8 hours**. When a token expires you will receive a `401`
with `"token has expired"` — re-authenticate to get a new one.

---

### POST /api/v1/auth/login

Authenticate with your AD username and password. Returns a signed JWT.

> **Public — no token required.**

#### Request body

```json
{
  "username": "jdoe",
  "password": "your-password"
}
```

| Field      | Type   | Required | Description             |
|------------|--------|----------|-------------------------|
| `username` | string | yes      | AD username (no domain) |
| `password` | string | yes      | AD password             |

#### Response `200 OK`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-05-19T06:00:00Z",
  "username": "jdoe",
  "display_name": "John Doe",
  "email": "jdoe@ecg.com.gh"
}
```

| Field          | Type   | Description                               |
|----------------|--------|-------------------------------------------|
| `token`        | string | JWT to include on all subsequent requests |
| `expires_at`   | string | ISO 8601 UTC expiry time                  |
| `username`     | string | AD username                               |
| `display_name` | string | Full display name from AD                 |
| `email`        | string | Email address from AD                     |

#### Error responses

| Status | Meaning                             |
|--------|-------------------------------------|
| `400`  | Missing username or password        |
| `401`  | Invalid username or password        |
| `500`  | Could not reach the LDAP server     |

---

### GET /api/v1/auth/me

Returns the identity and claims embedded in the current token. Useful for
verifying a token is still valid and reading the authenticated user's details.

**Requires:** `Authorization: Bearer <token>`

#### Response `200 OK`

```json
{
  "username": "jdoe",
  "display_name": "John Doe",
  "email": "jdoe@ecg.com.gh",
  "groups": [
    "CN=ECG-Staff,OU=Groups,DC=ecg,DC=com,DC=gh"
  ],
  "expires_at": "2026-05-19T06:00:00Z"
}
```

---

## Customers

Each customer response includes their **most recent tamper event** (if any).
When the customer record has no coordinates, the API automatically promotes
the event's coordinates so the map always has a point to render.

---

### GET /api/v1/customers

Returns a paginated list of customers. Supports rich filtering — see
[Filter Reference](#filter-reference) for all available parameters.

**Requires:** `Authorization: Bearer <token>`

#### Query parameters

| Parameter              | Type    | Description                                                          |
|------------------------|---------|----------------------------------------------------------------------|
| `search`               | string  | Free-text search across full name, account number, and meter number  |
| `region_code`          | string  | Exact match on region code (e.g. `01`)                               |
| `region_name`          | string  | Partial, case-insensitive match on region name                       |
| `district_code`        | string  | Exact match on district code (e.g. `01-06`)                          |
| `district_name`        | string  | Partial, case-insensitive match on district name                     |
| `service_type`         | string  | Exact — `Prepaid` or `Postpaid`                                      |
| `service_class`        | string  | Exact — `Residential` or `NonResidential`                            |
| `service_point_number` | string  | Exact match on service point number                                  |
| `customer_type`        | string  | Exact — `Individual` or `Organization`                               |
| `contract_status`      | string  | Exact — `Active` or `Inactive`                                       |
| `cms_contract_status`  | string  | Exact — `Active Contract`, `PendingReplacement`, etc.                |
| `meter_make`           | string  | Partial, case-insensitive match on meter make (e.g. `wasion`)        |
| `meter_model`          | string  | Partial, case-insensitive match on meter model                       |
| `account_type`         | string  | Exact match on account type (e.g. `NSLT`)                            |
| `has_tamper`           | boolean | `true` — only return customers who have at least one tamper event    |
| `page`                 | integer | Page number, 1-based. Default: `1`                                   |
| `page_size`            | integer | Records per page. Default: `50`. Maximum: `200`                      |

#### Example requests

```
# All prepaid residential customers in Accra East
GET /api/v1/customers?region_code=01&service_type=Prepaid&service_class=Residential

# All customers with tamper events in Kwabenya district
GET /api/v1/customers?district_name=kwabenya&has_tamper=true

# Search by name, page 2
GET /api/v1/customers?search=salifu&page=2&page_size=25

# All MMS-WASION meters with active contracts
GET /api/v1/customers?meter_make=wasion&contract_status=Active

# Organization accounts pending replacement
GET /api/v1/customers?customer_type=Organization&cms_contract_status=PendingReplacement
```

#### Response `200 OK`

```json
{
  "data": [
    {
      "ecgkey": "64767481d3388f971dc6667670000905720000091820000091820230530T221113110Z",
      "account_number": "700009057",
      "service_point_number": "200000918",
      "full_name": "SALIFU IDDRISU",
      "phone_number": "",
      "service_type": "Postpaid",
      "service_class": "NonResidential",
      "tariff_class": "Non - Residential",
      "contract_status": "Active",
      "meter_number": "0261120549",
      "meter_type": "Small",
      "meter_make": "Undefined 1 (Conv.)",
      "meter_model": "Undefined",
      "address": "6110-ACCRA, St1 - Kwabenya - MADINA ZONGO, MADINA ZONGO, KWABENYA",
      "street_name": "St1 - Kwabenya - MADINA ZONGO",
      "house_number": "6110-ACCRA",
      "region_name": "Accra East",
      "district_name": "Kwabenya",
      "latitude": null,
      "longitude": null,
      "last_bill_amount": 252.35,
      "current_balance": 3066.24,
      "last_payment_date": "2025-12-24",
      "last_payment_amount": 500.0,
      "latest_tamper_event": {
        "event_code": "3.26.38.98",
        "event_desc": "STDEVT:Power Up",
        "event_time": "2026-04-07T06:41:28Z",
        "period": "2026-04-07 06:04:28",
        "latitude": 5.6877487,
        "longitude": -0.0652943
      }
    }
  ],
  "total": 1240,
  "page": 1,
  "page_size": 50,
  "total_pages": 25
}
```

> **Note on coordinates:** `latitude` and `longitude` at the top level may be
> `null` in the raw database record. When that is the case and the tamper event
> has coordinates, the API promotes the event coordinates up automatically.
> Always use the top-level `latitude`/`longitude` fields when plotting on the map.

---

### GET /api/v1/customers/account/{accountNumber}

Returns a single customer by account number, with their latest tamper event.

**Requires:** `Authorization: Bearer <token>`

#### Path parameter

| Parameter       | Description                                |
|-----------------|--------------------------------------------|
| `accountNumber` | Customer account number (e.g. `700009057`) |

#### Example

```
GET /api/v1/customers/account/700009057
```

#### Response `200 OK`

Single customer object — same shape as one item from the `data` array above.

#### Error responses

| Status | Meaning           |
|--------|-------------------|
| `404`  | Account not found |

---

### GET /api/v1/customers/meter/{meterNumber}

Returns a single customer by meter number, with their latest tamper event.

**Requires:** `Authorization: Bearer <token>`

#### Path parameter

| Parameter     | Description                           |
|---------------|---------------------------------------|
| `meterNumber` | Meter number (e.g. `0261120549`)      |

#### Example

```
GET /api/v1/customers/meter/0261120549
```

#### Response `200 OK`

Single customer object — same shape as one item from the `data` array above.

#### Error responses

| Status | Meaning         |
|--------|-----------------|
| `404`  | Meter not found |

---

## Tamper Events

These endpoints return the **full event history** for a meter, newest first.
Both endpoints accept identical filter and pagination parameters.

---

### GET /api/v1/customers/meter/{meterNumber}/events

All tamper events for a specific meter number.

**Requires:** `Authorization: Bearer <token>`

#### Path parameter

| Parameter     | Description  |
|---------------|--------------|
| `meterNumber` | Meter number |

#### Query parameters

| Parameter    | Type    | Description                                               |
|--------------|---------|-----------------------------------------------------------|
| `event_desc` | string  | Partial, case-insensitive match on event description      |
| `event_code` | string  | Exact match on event code (e.g. `3.26.38.98`)             |
| `from`       | string  | Return events at or after this datetime (ISO 8601)        |
| `to`         | string  | Return events at or before this datetime (ISO 8601)       |
| `page`       | integer | Page number, 1-based. Default: `1`                        |
| `page_size`  | integer | Records per page. Default: `50`. Maximum: `500`           |

**Accepted datetime formats for `from` and `to`:**

```
2026-04-01T00:00:00Z       UTC with timezone — recommended
2026-04-01T00:00:00        No timezone — treated as UTC
2026-04-01                 Date only — time defaults to 00:00:00 UTC
```

#### Example requests

```
# All events for a meter in April 2026
GET /api/v1/customers/meter/24210186771/events?from=2026-04-01&to=2026-04-30

# Only "Power Up" events
GET /api/v1/customers/meter/24210186771/events?event_desc=Power+Up

# Exact event code within a date window
GET /api/v1/customers/meter/24210186771/events?event_code=3.26.38.98&from=2026-05-11T00:00:00Z

# Combine all filters
GET /api/v1/customers/meter/24210186771/events?event_desc=tamper&from=2026-01-01&page=1&page_size=100
```

#### Response `200 OK`

```json
{
  "meter_number": "24210186771",
  "account_number": "",
  "total": 38,
  "page": 1,
  "page_size": 50,
  "total_pages": 1,
  "events": [
    {
      "period": "2026-04-16 11:04:49",
      "meter_number": "24210186771",
      "customer_name": "BAWA AHMED",
      "event_code": "3.26.38.98",
      "event_desc": "STDEVT:Power Up",
      "event_time": "2026-04-16T11:28:49Z",
      "latitude": 5.6877487,
      "longitude": -0.0652943,
      "counting": null
    }
  ]
}
```

#### Error responses

| Status | Meaning                             |
|--------|-------------------------------------|
| `400`  | Invalid `from` or `to` datetime     |

---

### GET /api/v1/customers/account/{accountNumber}/events

All tamper events for the meter belonging to a given account number.
The meter number resolved from the account is included in the response.

**Requires:** `Authorization: Bearer <token>`

#### Path parameter

| Parameter       | Description             |
|-----------------|-------------------------|
| `accountNumber` | Customer account number |

Accepts the **same query parameters** as the meter events endpoint
(`event_desc`, `event_code`, `from`, `to`, `page`, `page_size`).

#### Example requests

```
# All events for an account in a date window
GET /api/v1/customers/account/700009064/events?from=2026-04-01&to=2026-04-30

# Power Up events only
GET /api/v1/customers/account/700009064/events?event_desc=Power+Up&page_size=200
```

#### Response `200 OK`

Same shape as the meter events response. `account_number` will be populated.

```json
{
  "meter_number": "24210186771",
  "account_number": "700009064",
  "total": 12,
  "page": 1,
  "page_size": 50,
  "total_pages": 1,
  "events": [ ... ]
}
```

#### Error responses

| Status | Meaning                             |
|--------|-------------------------------------|
| `400`  | Invalid `from` or `to` datetime     |
| `404`  | Account not found                   |

---

## Health Check

### GET /health

Returns server status. **Public — no token required.**

#### Response `200 OK`

```json
{ "status": "ok" }
```

---

## Filter Reference

### Customer filters

| Parameter              | Match   | Notes                                                   |
|------------------------|---------|---------------------------------------------------------|
| `search`               | partial | Searches full name, account number, and meter number    |
| `region_code`          | exact   | e.g. `01`, `05`                                         |
| `region_name`          | partial | e.g. `accra`, `eastern`                                 |
| `district_code`        | exact   | e.g. `01-06`, `05-09`                                   |
| `district_name`        | partial | e.g. `kwabenya`, `akim oda`                             |
| `service_type`         | exact   | `Prepaid` or `Postpaid`                                 |
| `service_class`        | exact   | `Residential` or `NonResidential`                       |
| `service_point_number` | exact   | e.g. `200000918`                                        |
| `customer_type`        | exact   | `Individual` or `Organization`                          |
| `contract_status`      | exact   | `Active` or `Inactive`                                  |
| `cms_contract_status`  | exact   | `Active Contract`, `Active`, `PendingReplacement`, etc. |
| `meter_make`           | partial | e.g. `wasion`, `gamma`, `landis`                        |
| `meter_model`          | partial | e.g. `ddsd101`, `cl710`                                 |
| `account_type`         | exact   | e.g. `NSLT`                                             |
| `has_tamper`           | boolean | `true` only — omit or set `false` to include all        |
| `page`                 | —       | Default `1`                                             |
| `page_size`            | —       | Default `50`, max `200`                                 |

### Event filters

| Parameter    | Match | Notes                                                    |
|--------------|-------|----------------------------------------------------------|
| `event_desc` | partial | e.g. `Power Up`, `tamper`, `magnetic`                  |
| `event_code` | exact   | e.g. `3.26.38.98`                                      |
| `from`       | >=    | Inclusive lower bound on `event_time`                    |
| `to`         | <=    | Inclusive upper bound on `event_time`                    |
| `page`       | —     | Default `1`                                              |
| `page_size`  | —     | Default `50`, max `500`                                  |

**Partial** matches are case-insensitive and match anywhere in the field value.  
**Exact** matches are case-sensitive — use the value exactly as it appears in the data.  
All filters are optional and can be combined freely.

---

## Response Shapes

### Customer object

```typescript
{
  ecgkey:               string
  account_number:       string
  service_point_number: string
  full_name:            string
  phone_number:         string
  service_type:         "Prepaid" | "Postpaid"
  service_class:        "Residential" | "NonResidential"
  tariff_class:         string
  contract_status:      "Active" | "Inactive"
  meter_number:         string
  meter_type:           string
  meter_make:           string
  meter_model:          string
  address:              string
  street_name:          string
  house_number:         string
  region_name:          string
  district_name:        string
  latitude:             number | null
  longitude:            number | null
  last_bill_amount:     number | null
  current_balance:      number | null
  last_payment_date:    string | null   // "YYYY-MM-DD"
  last_payment_amount:  number | null
  latest_tamper_event:  TamperEventSummary | null
}
```

### TamperEventSummary (nested inside customer)

```typescript
{
  event_code:  string
  event_desc:  string
  event_time:  string   // ISO 8601
  period:      string
  latitude:    number | null
  longitude:   number | null
}
```

### TamperEvent (inside events list)

```typescript
{
  period:        string
  meter_number:  string
  customer_name: string
  event_code:    string
  event_desc:    string
  event_time:    string   // ISO 8601
  latitude:      number | null
  longitude:     number | null
  counting:      number | null
}
```

### Paginated list wrapper

```typescript
{
  data:        T[]
  total:       number   // total records matching the current filters
  page:        number
  page_size:   number
  total_pages: number
}
```

### Events envelope

```typescript
{
  meter_number:   string
  account_number: string   // empty string when queried by meter
  total:          number
  page:           number
  page_size:      number
  total_pages:    number
  events:         TamperEvent[]
}
```

---

## Error Responses

All errors follow this shape:

```json
{
  "error":   "short description",
  "message": "detailed message (only present on 5xx errors)"
}
```

| Status | Meaning                                                   |
|--------|-----------------------------------------------------------|
| `400`  | Bad request — missing required field or invalid parameter |
| `401`  | Missing token, invalid token, or expired token            |
| `404`  | Requested resource not found                              |
| `500`  | Internal server error — check `message` for detail        |

### 401 bodies

```json
{ "error": "missing or malformed Authorization header" }
{ "error": "token has expired" }
{ "error": "invalid token" }
```

---

## MapLibre Integration

### Coordinate handling

Coordinates come from two sources in priority order:

1. The customer record's own `latitude` / `longitude`
2. The latest tamper event's `latitude` / `longitude`

The API falls back to source 2 automatically when source 1 is null.
Always use the top-level `latitude` / `longitude` fields on the customer object.

### Building a GeoJSON source

MapLibre expects `[longitude, latitude]` order (GeoJSON standard):

```js
const geojson = {
  type: "FeatureCollection",
  features: customers
    .filter(c => c.latitude !== null && c.longitude !== null)
    .map(c => ({
      type: "Feature",
      geometry: {
        type: "Point",
        coordinates: [c.longitude, c.latitude]  // longitude first
      },
      properties: {
        account_number:  c.account_number,
        full_name:       c.full_name,
        meter_number:    c.meter_number,
        service_type:    c.service_type,
        contract_status: c.contract_status,
        current_balance: c.current_balance,
        event_desc:      c.latest_tamper_event?.event_desc ?? null,
        event_time:      c.latest_tamper_event?.event_time ?? null,
      }
    }))
}

map.getSource("customers").setData(geojson)
```

### Fetching all pages

```js
async function fetchAllCustomers(token, filters = {}) {
  const params = new URLSearchParams({ page_size: 200, ...filters })
  let page = 1
  let all = []

  while (true) {
    params.set("page", page)
    const res = await fetch(`/api/v1/customers?${params}`, {
      headers: { Authorization: `Bearer ${token}` }
    })

    if (res.status === 401) {
      // Token expired — redirect to login
      throw new Error("token_expired")
    }

    const json = await res.json()
    all = all.concat(json.data)

    if (page >= json.total_pages) break
    page++
  }

  return all
}
```

### Handling token expiry

```js
async function apiRequest(url, token) {
  const res = await fetch(url, {
    headers: { Authorization: `Bearer ${token}` }
  })

  if (res.status === 401) {
    const body = await res.json()
    if (body.error === "token has expired") {
      // Clear stored token and redirect to login
      localStorage.removeItem("token")
      window.location.href = "/login"
    }
  }

  return res
}
```

---

## Tile Endpoints

Tile endpoints are **public within the API** but require a valid JWT passed as a `?token=` query parameter. This is necessary because MapLibre cannot send custom headers on tile requests.

### Available tile sources

| Source | Description | Best zoom |
|--------|-------------|-----------|
| `CustomerRecords` | All meters with coordinates | 7–20 |
| `customer_map_view` | Meters with latest event data attached | 7–20 |
| `district_event_summary` | District-level aggregates | 4–10 |

### GET /api/v1/tiles/{source}/{z}/{x}/{y}

Proxies MVT tiles from the Martin tile server after JWT validation.

**Auth:** `?token=<jwt>` query parameter (Bearer header also accepted)

**Path parameters:**

| Parameter | Description |
|-----------|-------------|
| `source` | Tile source name — see table above |
| `z` | Zoom level |
| `x` | Tile X coordinate |
| `y` | Tile Y coordinate |

**Response codes:**

| Code | Meaning |
|------|---------|
| `200` | Tile contains features — binary protobuf data |
| `204` | Tile valid but empty (no features at this location/zoom) |
| `401` | Missing or invalid token |
| `404` | Unknown tile source |
| `502` | Martin tile server unreachable |

---

## MapLibre Integration

### Adding a tile source

```js
// After login, store the token
const token = loginResponse.token

// Add customer meters as a vector tile source
map.addSource('meters', {
  type: 'vector',
  tiles: [
    `http://your-api-host:9400/api/v1/tiles/customer_map_view/{z}/{x}/{y}?token=${token}`
  ],
  minzoom: 5,
  maxzoom: 20
})

// Add a circle layer for individual meters (high zoom)
map.addLayer({
  id: 'meters-points',
  type: 'circle',
  source: 'meters',
  'source-layer': 'customer_map_view',
  minzoom: 10,
  paint: {
    'circle-radius': 5,
    'circle-color': [
      'match', ['get', 'contractstatus'],
      'Active', '#22c55e',
      '#ef4444'
    ]
  }
})

// Add district summary layer (low zoom overview)
map.addSource('districts', {
  type: 'vector',
  tiles: [
    `http://your-api-host:9400/api/v1/tiles/district_event_summary/{z}/{x}/{y}?token=${token}`
  ],
  minzoom: 4,
  maxzoom: 10
})
```

### Token expiry handling

MapLibre caches tile sources. When the JWT expires you need to update the tile URL:

```js
function refreshTileSource(map, token) {
  // Remove and re-add the source with the new token
  if (map.getSource('meters')) {
    map.removeLayer('meters-points')
    map.removeSource('meters')
  }
  map.addSource('meters', {
    type: 'vector',
    tiles: [`http://your-api-host:9400/api/v1/tiles/customer_map_view/{z}/{x}/{y}?token=${token}`]
  })
  // Re-add layers...
}
```

### Source layer names

The `source-layer` in MapLibre must match the table/view name exactly:

| Tile source | source-layer value |
|-------------|-------------------|
| `CustomerRecords` | `CustomerRecords` |
| `customer_map_view` | `customer_map_view` |
| `district_event_summary` | `district_event_summary` |

### Available properties per source

**`customer_map_view`**
```
ecgkey, meternumber, accountnumber, fullname,
servicetype, serviceclass, contractstatus, cmscontractstatus,
regioncode, regionname, districtcode, districtname,
metermake, metermodel, metertype, has_any_coordinates,
latest_event_time, latest_event_code, latest_event_desc, event_count
```

**`district_event_summary`**
```
districtcode, districtname, regionname,
total_meters, meters_with_events, meters_without_coordinates,
latest_event_time
```

### Ghana tile coordinates reference

| Area | Zoom | x | y |
|------|------|---|---|
| All Ghana | 6 | 31 | 31 |
| Accra area | 7 | 63 | 61 |
| Accra detail | 8 | 127 | 123 |
| Kumasi area | 7 | 63 | 61 |