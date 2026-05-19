# ─────────────────────────────────────────────────────────────────────────────
# ECG Backend — Makefile
# ─────────────────────────────────────────────────────────────────────────────

BINARY     = ecg-backend
MAIN       = ./cmd/api
BUILD_DIR  = bin
IMAGE_NAME = ecg-backend

# Colours for terminal output
RESET  = \033[0m
BOLD   = \033[1m
GREEN  = \033[32m
YELLOW = \033[33m
CYAN   = \033[36m

# ── Load .env file if it exists ───────────────────────────────────────────────
# Makes all vars in .env available to every make target without needing
# to manually export them in the shell first.
-include .env
export

# ── Build DATABASE_URL from parts if not explicitly set ──────────────────────
# Lets make migrate / make psql work straight from .env with no extra setup.
DATABASE_URL ?= postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.DEFAULT_GOAL := help


# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help
help:
	@echo ""
	@echo "$(BOLD)ECG Backend$(RESET)"
	@echo ""
	@echo "$(CYAN)Development$(RESET)"
	@echo "  make run            Run the API server locally (infra must be up)"
	@echo "  make infra-up        Start db + valkey, run migrations, then start martin"
	@echo "  make infra-down      Stop db, valkey, martin"
	@echo "  make build          Compile binary to $(BUILD_DIR)/$(BINARY)"
	@echo "  make tidy           Run go mod tidy"
	@echo "  make lint           Run golangci-lint"
	@echo "  make test           Run tests"
	@echo ""
	@echo "$(CYAN)Database$(RESET)"
	@echo "  make migrate        Run all pending migrations in order"
	@echo "  make migrate-000    Install extensions (PostGIS, TimescaleDB) — run first"
	@echo "  make migrate-001    Create tables and base indexes"
	@echo "  make migrate-002    Add coordinate flag columns + triggers"
	@echo "  make migrate-003    TimescaleDB hypertable for events"
	@echo "  make migrate-004    PostGIS geometry, meter_summary, Martin views"
	@echo "  make db-indexes     Print index health report"
	@echo "  make db-stats       Print table row counts and sizes"
	@echo ""
	@echo "$(CYAN)ETL$(RESET)"
	@echo "  make etl customers=<path> events=<path>   Load CSVs into staging DB"
	@echo ""
	@echo "$(CYAN)Docker$(RESET)"
	@echo "  make up             Start all services (db, martin, api)"
	@echo "  make down           Stop all services"
	@echo "  make up-db          Start database only"
	@echo "  make logs           Tail all service logs"
	@echo "  make logs-api       Tail API logs"
	@echo "  make logs-db        Tail DB logs"
	@echo "  make logs-martin    Tail Martin tile server logs
	@echo "  make logs-valkey     Tail Valkey logs""
	@echo "  make docker-build   Build the API Docker image"
	@echo "  make restart-api    Rebuild and restart the API container only"
	@echo ""
	@echo "$(CYAN)Utilities$(RESET)"
	@echo "  make psql           Open a psql shell to the staging DB
	@echo "  make valkey-cli        Open a valkey-cli shell"
	@echo ""
	@echo "$(CYAN)Utilities$(RESET)"
	@echo "  make hash-password password=<pw>   Generate bcrypt hash for dev account""
	@echo "  make clean          Remove build artifacts"
	@echo ""
	@echo "$(YELLOW)DATABASE_URL must be set for all db/etl targets.$(RESET)"
	@echo "$(YELLOW)Use 'cp .env.example .env' and fill in your values.$(RESET)"
	@echo ""

# ── Development ───────────────────────────────────────────────────────────────

.PHONY: run
run:
	@echo "$(GREEN)Starting API server...$(RESET)"
	go run $(MAIN)

# Start only the backing services (db, valkey, martin) without the API container.
# Used by 'make run' so the Go process runs locally against live infra.
.PHONY: infra-up
infra-up:
	@echo "$(GREEN)Starting backing services (db, valkey)...$(RESET)"
	$(COMPOSE) up -d db valkey
	@echo "$(CYAN)Waiting for DB to be healthy...$(RESET)"
	@until $(COMPOSE) exec -T db pg_isready -U $(DB_USER) -d $(DB_NAME) -p $(DB_PORT) -q; do sleep 1; done
#	@echo "$(CYAN)Running migrations...$(RESET)"
#	@$(MAKE) migrate
	@echo "$(GREEN)Starting Martin tile server...$(RESET)"
	$(COMPOSE) up -d martin
	@echo "$(GREEN)All backing services up$(RESET)"

.PHONY: infra-down
infra-down:
	@echo "$(YELLOW)Stopping backing services...$(RESET)"
	$(COMPOSE) stop db valkey martin

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	@echo "$(GREEN)Building $(BINARY)...$(RESET)"
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) $(MAIN)
	@echo "$(GREEN)Built: $(BUILD_DIR)/$(BINARY)$(RESET)"

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./... -v -race -count=1

.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Cleaned$(RESET)"

# ── Database migrations ───────────────────────────────────────────────────────

# Build explicit psql connection flags from individual vars — more reliable
# on Windows/Git Bash than passing a full DATABASE_URL string.
# PGPASSWORD is read automatically by psql — avoids interactive password prompts.
export PGPASSWORD = $(DB_PASSWORD)
PSQL = psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

_check-db-url:
	@test -n "$(DB_HOST)" || (echo "$(YELLOW)ERROR: DB_HOST is not set — check your .env$(RESET)" && exit 1)
	@test -n "$(DB_PASSWORD)" || (echo "$(YELLOW)ERROR: DB_PASSWORD is not set — check your .env$(RESET)" && exit 1)

.PHONY: migrate
migrate: _check-db-url
	@echo "$(GREEN)Running all migrations...$(RESET)"
	@$(MAKE) migrate-000
	@$(MAKE) migrate-001
	@$(MAKE) migrate-002
	@$(MAKE) migrate-003
	@$(MAKE) migrate-004
	@echo "$(GREEN)All migrations complete$(RESET)"

.PHONY: migrate-000
migrate-000: _check-db-url
	@echo "$(CYAN)000 — extensions (PostGIS, TimescaleDB)$(RESET)"
	$(PSQL) -f migrations/000_extensions.sql

.PHONY: migrate-001
migrate-001: _check-db-url
	@echo "$(CYAN)001 — tables and base indexes$(RESET)"
	$(PSQL) -f migrations/001_create_tables.sql

.PHONY: migrate-002
migrate-002: _check-db-url
	@echo "$(CYAN)002 — coordinate flag columns and triggers$(RESET)"
	$(PSQL) -f migrations/002_coordinate_flags.sql

.PHONY: migrate-003
migrate-003: _check-db-url
	@echo "$(CYAN)003 — TimescaleDB hypertable for events$(RESET)"
	$(PSQL) -f migrations/003_extensions_and_hypertable.sql

.PHONY: migrate-004
migrate-004: _check-db-url
	@echo "$(CYAN)004 — PostGIS geometry, meter_summary, Martin views$(RESET)"
	$(PSQL) -f migrations/004_postgis_and_summary.sql

.PHONY: post-load
post-load: _check-db-url
	@echo "$(GREEN)Running post-load refresh...$(RESET)"
	$(PSQL) -f etl/post_load.sql

# Force a full re-run of post-load from step 1 (ignores checkpoints)
.PHONY: post-load-reset
post-load-reset: _check-db-url
	@echo "$(YELLOW)Resetting post-load checkpoints — next run will start from step 1$(RESET)"
	@$(PSQL) -c "SELECT post_load_reset();"
	@echo "$(GREEN)Checkpoints cleared — run 'make post-load' to start fresh$(RESET)"

.PHONY: db-reset
db-reset: _check-db-url
	@echo "$(YELLOW)Dropping and recreating all tables...$(RESET)"
	@$(PSQL) -c 'DROP TABLE IF EXISTS post_load_checkpoint CASCADE;'
	@$(PSQL) -c 'DROP TABLE IF EXISTS district_event_summary CASCADE;'
	@$(PSQL) -c 'DROP TABLE IF EXISTS meter_summary CASCADE;'
	@$(PSQL) -c 'DROP TABLE IF EXISTS "MMS_METER_TAMPER_EVENTS" CASCADE;'
	@$(PSQL) -c 'DROP TABLE IF EXISTS "CustomerRecords" CASCADE;'
	@$(PSQL) -c 'DROP VIEW IF EXISTS customer_map_view CASCADE;'
	@$(PSQL) -c 'DROP VIEW IF EXISTS account_meters CASCADE;'
	@echo "$(GREEN)Tables dropped — running migrations...$(RESET)"
	@$(MAKE) migrate

.PHONY: add-indexes
add-indexes: _check-db-url
	@echo "$(CYAN)Adding missing indexes to running DB...$(RESET)"
	@$(PSQL) -c 'CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cr_servicetype ON "CustomerRecords" (servicetype);'
	@$(PSQL) -c 'CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cr_contractstatus ON "CustomerRecords" (contractstatus);'
	@echo "$(CYAN)Adding service stat columns to district_event_summary...$(RESET)"
	@$(PSQL) -c 'ALTER TABLE district_event_summary ADD COLUMN IF NOT EXISTS prepaid_meters BIGINT NOT NULL DEFAULT 0;'
	@$(PSQL) -c 'ALTER TABLE district_event_summary ADD COLUMN IF NOT EXISTS postpaid_meters BIGINT NOT NULL DEFAULT 0;'
	@$(PSQL) -c 'ALTER TABLE district_event_summary ADD COLUMN IF NOT EXISTS active_contracts BIGINT NOT NULL DEFAULT 0;'
	@echo "$(GREEN)Done — run make post-load-reset && make post-load to repopulate$(RESET)"

.PHONY: db-stats
db-stats: _check-db-url
	@echo "$(CYAN)Table row counts and sizes$(RESET)"
	@$(PSQL) -c "\
		SELECT \
			relname AS table, \
			to_char(n_live_tup, '999,999,999') AS rows, \
			pg_size_pretty(pg_total_relation_size(quote_ident(relname))) AS total_size, \
			pg_size_pretty(pg_relation_size(quote_ident(relname))) AS table_size, \
			pg_size_pretty(pg_total_relation_size(quote_ident(relname)) \
				- pg_relation_size(quote_ident(relname))) AS index_size \
		FROM pg_stat_user_tables \
		WHERE relname IN ( \
			'CustomerRecords', \
			'MMS_METER_TAMPER_EVENTS', \
			'meter_summary', \
			'district_event_summary' \
		) \
		ORDER BY pg_total_relation_size(quote_ident(relname)) DESC;"

.PHONY: db-indexes
db-indexes: _check-db-url
	@echo "$(CYAN)Index usage report$(RESET)"
	@$(PSQL) -c "\
		SELECT \
			relname AS table, \
			indexrelname AS index, \
			idx_scan AS scans, \
			pg_size_pretty(pg_relation_size(indexrelid)) AS size \
		FROM pg_stat_user_indexes \
		WHERE relname IN ( \
			'CustomerRecords', \
			'MMS_METER_TAMPER_EVENTS', \
			'meter_summary', \
			'district_event_summary' \
		) \
		ORDER BY relname, idx_scan DESC;"

# ── ETL ───────────────────────────────────────────────────────────────────────

.PHONY: etl
etl: _check-db-url
	@test -n "$(customers)" || (echo "$(YELLOW)Usage: make etl customers=<path> events=<path>$(RESET)" && exit 1)
	@test -n "$(events)"    || (echo "$(YELLOW)Usage: make etl customers=<path> events=<path>$(RESET)" && exit 1)
	@echo "$(GREEN)Running ETL load...$(RESET)"
	DATABASE_URL="$$DATABASE_URL" ./etl/run_etl.sh $(customers) $(events)

# ── Docker ────────────────────────────────────────────────────────────────────

COMPOSE = docker compose -f deploy/docker-compose.yml --env-file .env

.PHONY: up
up:
	@echo "$(GREEN)Starting all services...$(RESET)"
	$(COMPOSE) up -d
	@echo "$(GREEN)Services running:$(RESET)"
	@$(COMPOSE) ps

.PHONY: up-db
up-db:
	@echo "$(GREEN)Starting database...$(RESET)"
	$(COMPOSE) up -d db
	@echo "$(GREEN)Waiting for DB to be healthy...$(RESET)"
	@$(COMPOSE) exec db pg_isready -U supabase_admin -d ecg

.PHONY: down
down:
	@echo "$(YELLOW)Stopping all services...$(RESET)"
	$(COMPOSE) down

.PHONY: logs
logs:
	$(COMPOSE) logs -f

.PHONY: logs-api
logs-api:
	$(COMPOSE) logs -f api

.PHONY: logs-db
logs-db:
	$(COMPOSE) logs -f db

.PHONY: logs-martin
logs-martin:
	$(COMPOSE) logs -f martin

.PHONY: logs-valkey
logs-valkey:
	$(COMPOSE) logs -f valkey

.PHONY: docker-build
docker-build:
	@echo "$(GREEN)Building Docker image $(IMAGE_NAME)...$(RESET)"
	docker build -t $(IMAGE_NAME):latest .

.PHONY: docker-run
docker-run:
	@echo "$(GREEN)Starting full stack (all services including API container)...$(RESET)"
	$(COMPOSE) up -d
	@echo "$(GREEN)Stack is up$(RESET)"
	@echo "  API:    http://localhost:9400"
	@echo "  Tiles:  http://localhost:9401"
	@echo "  DB:     localhost:9432"

.PHONY: docker-stop
docker-stop:
	@echo "$(YELLOW)Stopping all containers...$(RESET)"
	$(COMPOSE) down

.PHONY: docker-logs
docker-logs:
	$(COMPOSE) logs -f api

.PHONY: restart-api
restart-api:
	@echo "$(GREEN)Rebuilding and restarting API container...$(RESET)"
	$(COMPOSE) up -d --no-deps --build api

# ── Utilities ─────────────────────────────────────────────────────────────────

.PHONY: psql
psql: _check-db-url
	$(PSQL)

# ── Password hashing ──────────────────────────────────────────────────────────

.PHONY: hash-password
hash-password:
	@test -n "$(password)" || (echo "$(YELLOW)Usage: make hash-password password=yourpassword$(RESET)" && exit 1)
	@go run ./tools/hashpw/main.go "$(password)"

.PHONY: valkey-cli
valkey-cli:
	@echo "$(CYAN)Connecting to Valkey on port 9479...$(RESET)"
	docker exec -it ecg_valkey valkey-cli -p 9479