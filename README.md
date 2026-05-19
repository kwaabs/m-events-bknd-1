# ECG Backend

Go + Chi + Bun ORM backend for the ECG meter tamper map. Serves customer records
joined with their most recent tamper event, designed to feed a MapLibre frontend.

## Stack

| Layer    | Tech                                    |
|----------|-----------------------------------------|
| Language | Go 1.22                                 |
| Router   | [Chi v5](https://github.com/go-chi/chi) |
| ORM      | [Bun](https://bun.uptrace.dev/) + pgdriver |
| Database | PostgreSQL (tables already exist)       |

## Project Structure

```
github.com/kwaabs/m-events/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── config/config.go         # Env-based config
│   ├── database/database.go     # Bun DB setup
│   ├── models/models.go         # CustomerRecord, TamperEvent, DTOs
│   ├── repository/
│   │   ├── customer_repository.go  # All DB queries
│   │   └── helpers.go              # Type coercion helpers
│   ├── handlers/
│   │   └── customer_handler.go  # HTTP handlers
│   └── middleware/
│       └── logger.go            # Structured request logger
├── migrations/
│   └── 001_create_tables.sql    # Schema + indexes
├── .env.example
├── Dockerfile
└── Makefile
```

## Quick Start

```bash
# 1. Install dependencies
go mod tidy

# 2. Configure environment
cp .env.example .env
# Edit .env with your Postgres credentials

# 3. Apply migration (if tables don't exist yet)
make migrate

# 4. Run
make run
```

## API Endpoints

### Health
```
GET /health
```

### List Customers (with latest tamper event)
```
GET /api/v1/customers
```
Query parameters:

| Param           | Type    | Description                                  |
|-----------------|---------|----------------------------------------------|
| `search`        | string  | Free-text search on name, account, meter no. |
| `region`        | string  | Filter by `regioncode`                       |
| `district`      | string  | Filter by `districtcode`                     |
| `service_type`  | string  | `Prepaid` or `Postpaid`                      |
| `contract_status` | string | `Active`, `Inactive`, etc.                 |
| `has_tamper`    | bool    | `true` → only customers with tamper events   |
| `page`          | int     | Page number (default 1)                      |
| `page_size`     | int     | Records per page (default 50, max 200)       |

**Response:**
```json
{
  "data": [
    {
      "ecgkey": "...",
      "account_number": "700009057",
      "full_name": "SALIFU IDDRISU",
      "meter_number": "0261120549",
      "latitude": null,
      "longitude": null,
      "latest_tamper_event": {
        "event_code": "3.26.38.98",
        "event_desc": "STDEVT:Power Up",
        "event_time": "2026-04-07T06:41:28Z",
        "latitude": 5.68774870,
        "longitude": -0.06529430
      }
    }
  ],
  "total": 1240,
  "page": 1,
  "page_size": 50,
  "total_pages": 25
}
```

### Get Customer by Account Number
```
GET /api/v1/customers/account/{accountNumber}
```

### Get Customer by Meter Number
```
GET /api/v1/customers/meter/{meterNumber}
```

### Get All Tamper Events for a Meter
```
GET /api/v1/customers/meter/{meterNumber}/events?limit=100
```

## MapLibre Integration Notes

The `latest_tamper_event` object includes `latitude`/`longitude` from the event
itself. When `cr.latitude`/`cr.longitude` are null (common in this dataset), the
DTO automatically falls back to the event coordinates — so your map always has
a coordinate to plot if one exists anywhere in the record.

Suggested GeoJSON feature properties for a MapLibre source:
```js
{
  type: "Feature",
  geometry: {
    type: "Point",
    coordinates: [customer.longitude, customer.latitude] // lng, lat order
  },
  properties: {
    account_number: customer.account_number,
    full_name: customer.full_name,
    event_desc: customer.latest_tamper_event?.event_desc,
    event_time: customer.latest_tamper_event?.event_time,
  }
}
```

## Environment Variables

| Variable      | Default     | Description             |
|---------------|-------------|-------------------------|
| `PORT`        | `8080`      | HTTP listen port        |
| `APP_ENV`     | `development` | Environment name      |
| `DB_HOST`     | `localhost` | Postgres host           |
| `DB_PORT`     | `5432`      | Postgres port           |
| `DB_USER`     | `postgres`  | Postgres user           |
| `DB_PASSWORD` | —           | Postgres password       |
| `DB_NAME`     | `ecg`       | Database name           |
| `DB_SSLMODE`  | `disable`   | SSL mode                |
