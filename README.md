# GoFlow

Distributed job processing system in Go: worker pool, retries with exponential backoff, REST API, and metrics.

## Quick start (recommended: Docker Compose)

1. **Start Postgres** (creates user `goflow`, database `goflow`):

   ```bash
   docker compose up -d
   ```

   Wait until the postgres service is healthy (~5s). Optionally: `docker compose ps`.

2. **Configure env** — `.env` is not committed. Copy the example and edit if needed:

   ```bash
   cp .env.example .env
   ```

3. **Run the API** from the project root (migrations run on startup). Compose maps Postgres to **port 5433** so it doesn’t conflict with a local Postgres on 5432. Go does not load `.env` automatically—source it first:

   ```bash
   source .env && go run ./cmd/goflow
   ```
   Or: `export DATABASE_URL="postgres://goflow:goflow@localhost:5433/goflow?sslmode=disable"` then `go run ./cmd/goflow`.

   Server listens on `:8080`.

**Without Docker Compose:** use Postgres on 5432 (or your port), create user/db if needed, then set `DATABASE_URL` (e.g. `...@localhost:5432/goflow?...`) and run `go run ./cmd/goflow`.

## Test with Postman

1. Import the collection: **Postman → Import →** select `postman/GoFlow-API.postman_collection.json`.
2. Ensure the API is running (`go run ./cmd/goflow`).
3. Run **Health** to confirm the server and DB are up.
4. Run **Create Job (report)** or **Create Job (heavy_task)**; the collection will store the returned `id` in the variable `jobId`.
5. Run **Get Job by ID** to see status and result (pending → processing → completed).
6. Run **Get Metrics** to see job counts and average processing time.

Collection variable `baseUrl` defaults to `http://localhost:8080`; change it if your server runs elsewhere.

## API

- **GET /health** — Liveness: 200 if app and DB are OK, 503 if DB is down.
- **POST /jobs** — Submit a job. Body: `{"type": "report"|"image"|"email"|"heavy_task", "payload": {...}, "priority": 0}`. Response: `{"id": "...", "status": "pending", "created_at": "..."}`.
- **GET /jobs/:id** — Job status and result.
- **GET /metrics** — Counts and average processing time (JSON).

## Tests

```bash
go test ./...
```

## Config

- `DATABASE_URL` — PostgreSQL connection string. Use port **5433** when using Docker Compose (default host port); **5432** when using a local Postgres.
- `MIGRATIONS_DIR` — Path to migrations directory (default: `migrations`; run from project root).
- `WORKER_COUNT` — Number of worker goroutines (default: 3).

## Docs

See [docs/GOFLOW_BRIEF.md](docs/GOFLOW_BRIEF.md) for architecture, schema, and design notes.
