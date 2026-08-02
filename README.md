# NutriChat — AI Nutrition Assistant

A full-stack nutrition tracking app built to practice React/TypeScript, Go, PostgreSQL,
and AI application design. Users will (eventually) log meals in plain English —
e.g. *"I had 230g chicken thighs, 150g cooked rice, and a Quest protein shake"* —
and get back calculated calories, protein, carbs, fat, and fiber.

The core design principle: **the LLM never does nutrition math**. It only parses
natural language into structured data. All matching, unit conversion, and
calculation is deterministic Go code.

## Architecture

```text
React Website
   ↓
Go API Backend
   ↓
Nutrition Engine
   ├── Food matcher
   ├── Unit converter
   ├── Nutrition calculator
   ├── Ambiguity detector
   ├── Concurrent lookup service
   └── Meal logging service
   ↓
PostgreSQL
```

### Why Go owns the nutrition engine

Nutrition math needs to be correct and reproducible every time — an LLM
re-deriving calories from grams and macros is a source of silent, hard-to-catch
errors. Go handles food matching, unit conversion, and calculation as typed,
tested, deterministic logic. The LLM's job is limited to understanding free-form
text and extracting structured intent (foods, quantities, units) for Go to act on.

## Stack

- **Frontend:** React, TypeScript, Vite, Tailwind CSS
- **Backend:** Go, Chi
- **Database:** PostgreSQL (coming in Week 2 — currently in-memory seed data)
- **AI:** LLM structured-output parsing for natural language meal logging (Week 4)

## Project status

Currently in Week 1 (React + Go foundation). Implemented so far:

- Go domain models: `Food`, `Nutrition` ([backend/internal/nutrition/models.go](backend/internal/nutrition/models.go))
- In-memory seed food data ([backend/internal/foods/data.go](backend/internal/foods/data.go))
- Food search by substring match ([backend/internal/foods/search.go](backend/internal/foods/search.go))
- HTTP API via Chi ([backend/cmd/api/main.go](backend/cmd/api/main.go)):
  - `GET /health`
  - `GET /foods`
  - `GET /foods/search?q=`
  - `POST /nutrition/calculate`
- Nutrition calculation from food + quantity + unit ([backend/internal/nutrition/calculator.go](backend/internal/nutrition/calculator.go))
- React + TypeScript + Vite frontend scaffold ([frontend/](frontend/))

Not yet implemented: PostgreSQL persistence, unit conversion across multiple
unit types, meal logging, concurrent ingredient lookup, ambiguity detection,
AI parsing layer.

## Running locally

### Backend

```bash
cd backend
go run ./cmd/api
```

Server starts on `http://localhost:8080`.

```bash
curl http://localhost:8080/health
curl http://localhost:8080/foods
curl "http://localhost:8080/foods/search?q=chicken"
curl -X POST http://localhost:8080/nutrition/calculate \
  -H "Content-Type: application/json" \
  -d '{"food_name": "Chicken Thigh, Raw", "quantity": 230, "unit": "grams"}'
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Vite dev server starts on `http://localhost:5173`.
