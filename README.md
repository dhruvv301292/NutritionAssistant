# NutriChat — AI Nutrition Assistant

A full-stack nutrition tracking app built to practice React/TypeScript, Go, PostgreSQL,
and AI application architecture. Users log meals in plain English —
e.g. *"I had 230g chicken thighs, 150g cooked rice, and a Quest protein shake"* —
and get back calculated calories, protein, carbs, fat, and fiber, saved to a
daily log.

The core design principle: **the LLM never does nutrition math**. It only parses
natural language into structured data (foods, quantities, units). All food
matching, unit conversion, and calculation is deterministic Go code — the
same code path whether a meal comes in via the structured form or the AI
chat interface.

## Architecture

```text
React Website
   ↓
Go API Backend
   ↓
Nutrition Engine
   ├── Food matcher (Postgres trigram similarity + ambiguity detection)
   ├── Unit converter (grams / ounces / count, incl. count-based foods
   │   like eggs via a stored grams-per-unit reference)
   ├── Nutrition calculator (deterministic scaling from stored per-serving data)
   ├── Concurrent ingredient lookup (one goroutine per ingredient, errgroup)
   └── Meal logging service (save / fetch / daily summary)
   ↓
PostgreSQL
```

AI path:

```text
User natural language
   ↓
LLM parser (Claude, tool-use with an enum-constrained JSON schema)
   ↓
Structured meal JSON — { name, quantity, unit, preparation, brand }
   ↓
Go nutrition engine (identical matching/conversion/calculation as the
   non-AI path — the LLM's output is just another structured request)
   ↓
Calculated result + ambiguity/clarification flags
   ↓
User-friendly response (React chat UI)
```

### Why Go owns the nutrition engine

Nutrition math needs to be correct and reproducible every time — an LLM
re-deriving calories from grams and macros is a source of silent, hard-to-catch
errors that also can't be unit tested in any meaningful way. Go handles food
matching, unit conversion, and calculation as typed, tested, deterministic
logic (see `backend/internal/nutrition/*_test.go`,
`backend/internal/meals/matcher_test.go`). The LLM's job is strictly limited
to understanding free-form text and extracting structured intent — it never
sees or produces a calorie number. Both the manual meal-logging form and the
AI chat interface funnel into the exact same Go service (`meals.Service`), so
there's only one nutrition calculation code path to trust.

## Stack

- **Frontend:** React, TypeScript, Vite
- **Backend:** Go, Chi
- **Database:** PostgreSQL (Supabase), with `pg_trgm` for typo-tolerant food search
- **AI:** Anthropic Claude, structured tool-use output for natural-language meal parsing

## Project status

Weeks 1–4 of the build plan are complete (see `claude.md` for the full
day-by-day log). Implemented:

**Backend** ([backend/](backend/))
- Typed nutrition domain models and deterministic calculator
  ([internal/nutrition/](backend/internal/nutrition/))
- Unit conversion (grams ↔ ounces, plus count-based foods like eggs via a
  `grams_per_unit` reference) — [units.go](backend/internal/nutrition/units.go)
- Postgres-backed food repository/service with `pg_trgm` fuzzy search
  ([internal/foods/](backend/internal/foods/))
- Meal logging: concurrent multi-ingredient resolution via goroutines +
  `errgroup`, ambiguity detection (score-gap threshold between top matches),
  save/fetch, daily summaries computed from live food data (not snapshotted)
  — [internal/meals/](backend/internal/meals/)
- AI meal parsing via Claude tool-use with a JSON-schema-constrained output
  ([internal/ai/](backend/internal/ai/))
- HTTP API via Chi:
  - `GET /health`
  - `GET /foods`, `GET /foods/search?q=`
  - `POST /nutrition/calculate`
  - `POST /meals/calculate`, `POST /meals`, `GET /meals/today`, `GET /summary/daily`
  - `POST /ai/parse-meal`, `POST /chat/meal`
- Unit tests for conversion, calculation, ambiguity logic, and concurrent
  aggregation (`go test -race ./...` passes clean)

**Frontend** ([frontend/](frontend/))
- Food search with per-100g / per-serving nutrition display
- Structured meal logger (multi-ingredient form, live preview, save)
- Natural-language chat meal logger with nutrition cards and a clarification
  flow for ambiguous/unmatched items
- Daily dashboard (totals + meal history)
- Loading, error, and empty states across all of the above

**Not yet implemented:** authentication (a placeholder `user_id` is used
throughout), Docker, MCP server, external food database fallback for unseeded
foods, mobile app.

## Running locally

### Backend

Requires a `backend/.env` with `DATABASE_URL` (Postgres/Supabase connection
string) and `ANTHROPIC_API_KEY`.

```bash
cd backend
go run ./cmd/api
```

Server starts on `http://localhost:8080`.

```bash
curl http://localhost:8080/health

curl -X POST http://localhost:8080/chat/meal \
  -H "Content-Type: application/json" \
  -d '{"text": "I had 230g chicken thighs, 150g cooked rice, and a Quest protein shake"}'
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Vite dev server starts on `http://localhost:5173` and proxies `/api/*` to the
Go backend on `:8080`.

### Tests

```bash
cd backend
go test -race ./...
```
