# Claude Coaching Guide: AI Nutrition Assistant Website

## Project Overview

You are coaching me while I build an AI Nutrition Assistant as a resume-quality full-stack project.

The project will start as a website and later support an iPhone app using the same backend.

Core goals:

1. Learn React + TypeScript by building the web frontend.
2. Learn Go by building the backend API and nutrition engine.
3. Learning PostgreSQL is not a goal as I already know it. Still, advise me on best practices and inform me of design decisions regarding schema design and querying.
4. Learn AI application architecture by integrating an LLM for natural-language meal logging.
5. Optionally add MCP later as an AI-access layer for the nutrition backend.

The app should not be a generic chatbot. The LLM should parse natural language and generate user-friendly responses. The Go backend should own deterministic nutrition logic.

Target MVP:

A user can type: “I had 230g chicken thighs, 150g cooked rice, and a Quest protein shake.”

The app should:

1. Parse the meal.
2. Match foods against a nutrition database or seed data.
3. Convert quantities and units.
4. Calculate calories, protein, carbs, fat, and fiber.
5. Save the meal log.
6. Show a daily summary.
7. Display meal history in the UI.

---

## Stack

Frontend:
- React
- TypeScript
- Vite
- Tailwind CSS

Backend:
- Go
- Gin or Chi
- PostgreSQL

AI:
- OpenAI or Anthropic API
- Structured output / function-calling style parsing

Later:
- Docker
- MCP server
- React Native / Expo iPhone app
- Cloud deployment

---

## UI Philosophy

React UI structure, components, and layout are intentionally left open-ended.
Each frontend day below states the *functional* goal (what data/interactions
are needed), not a fixed component architecture. Specific components, layouts,
and styling decisions are made together during the session when we get there.

---

## Key Architecture Rule

Do not let the LLM do nutrition math.

LLM responsibilities:
- Parse natural language meal descriptions
- Extract foods, quantities, units, brands, and preparation notes
- Ask clarifying questions
- Generate friendly coaching responses

Go backend responsibilities:
- Normalize food names
- Match ingredients against food records
- Fall back to an external food database API when no local match exists, and
  cache the result in PostgreSQL for future lookups
- Handle raw vs cooked variants
- Convert units
- Run concurrent food lookups
- Calculate nutrition totals
- Detect ambiguity
- Save meal logs
- Return deterministic structured results

---

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

AI path:

```text
User natural language
   ↓
LLM parser
   ↓
Structured meal JSON
   ↓
Go nutrition engine
   ↓
Calculated result
   ↓
User-friendly response
```

---

## Food Lookup Flow (Cache-Aside Pattern)

When the nutrition engine needs data for a food, it follows a cache-aside pattern:

```text
1. Normalize the food name
2. Search PostgreSQL (foods, food_aliases)
   - Found     -> return matched food
   - Not found -> query external food database API
                  -> map response to internal Food model
                  -> insert into PostgreSQL (foods / nutrition_profiles)
                  -> return as matched food
```

This keeps PostgreSQL as the source of truth for anything the app has already
seen, while still being able to resolve foods that were never seeded.

Notes for later:
- Which external API to use (e.g., USDA FoodData Central) is decided when we
  build this — Week 2/3, not now.
- Concurrent requests for the same unseen food could race to insert it twice —
  handle with an upsert (`ON CONFLICT DO NOTHING`) when we get there.
- API credentials are managed via environment variables / Docker secrets
  (Phase 2).

---

## Backend Design Focus

The Go backend should highlight Go's strengths:

1. Typed domain models
2. Deterministic calculation logic
3. Concurrent ingredient lookup
4. Clean API boundaries
5. Repository/service pattern
6. Reliable error handling
7. Unit-tested nutrition logic

Suggested backend structure:

```text
backend/
  cmd/api/main.go
  internal/api/
  internal/nutrition/
    models.go
    calculator.go
    matcher.go
    units.go
    concurrent_lookup.go
  internal/foods/
  internal/fooddata/   # external food DB API client (cache-aside fallback)
  internal/meals/
  internal/ai/
  internal/db/
```

---

## Week 1: June 11–17 — React + Go Foundation

Goal:
- React app runs locally
- Go API runs locally
- React calls Go API
- Basic nutrition structs exist

Day 1:
- Create repo: `nutrichat`
- Create `frontend/` and `backend/`
- Learn Go structs and methods
- Build `Food` and `Nutrition` structs

Day 2:
- Learn slices and maps
- Create hardcoded food list
- Build `SearchFoods(query string) []Food`

Day 3:
- Learn Go HTTP basics
- Add Gin or Chi
- Build:
  - `GET /health`
  - `GET /foods`
  - `GET /foods/search?q=chicken`

Day 4:
- Create React + TypeScript + Vite app
- Build a homepage
- Sketch initial UI for searching foods and viewing nutrition info
  (component structure/names decided together during the session)

Day 5:
- Learn `useState`, `useEffect`, and `fetch`
- Connect React to Go API

Day 6:
- Build first nutrition calculator:
  - `CalculateNutrition(food Food, grams float64) Nutrition`
- Add `POST /nutrition/calculate`

Day 7:
- Refactor files
- Separate handlers from nutrition logic
- Add basic README notes

---

## Week 2: June 18–24 — PostgreSQL + Nutrition Engine

Goal:
- Food data moves from hardcoded Go data to Postgres
- Search and nutrition calculation use database-backed records

Day 8:
- Learn SQL basics
- Create Supabase project and database
- Create `foods` table

Day 9:
- Create:
  - `foods`
  - `food_aliases`
  - `nutrition_profiles`
- Seed common foods:
  - chicken thigh cooked
  - chicken breast cooked
  - white rice cooked
  - Greek yogurt
  - whey protein
  - egg
  - milk
  - homemade smash burger
  - Quest protein shake

Day 10:
- Connect Go to Postgres
- Replace hardcoded list with DB query

Day 11:
- Add repository/service pattern:
  - `FoodRepository`
  - `FoodService`

Day 12:
- Build food matching v1:
  1. Normalize query
  2. Exact match
  3. Alias match
  4. Partial match
  5. Account for spelling mistakes and typos (e.g. pg_trgm similarity)
  6. Ranked candidates
  6. If still no match: fall back to the external food database API, persist
     the result to PostgreSQL, and return it (see "Food Lookup Flow")

Day 13:
- Build unit conversion v1:
  - grams
  - ounces
  - serving
- Function:
  - `ConvertToGrams(quantity float64, unit string, food Food)`

Day 14:
- Extend unit conversion to handle units that don't match the food's native
  unit family (e.g. converting a count-based food like `egg` to/from grams)
- Add a `grams_per_unit` reference field (e.g. `foods` table column) for
  count-based foods so weight-based quantities become meaningful for them
- Update `ConvertToGrams` to use it when present, and fall back to
  exact-unit-match behavior when it isn't

Day 15:
- Improve React food search UI
- Show serving size and nutrition per 100g or per serving

---

## Week 3: June 25–July 1 — Meal Logging + Go Concurrency

Goal:
- User can log a meal
- Backend calculates totals
- Go resolves multiple ingredients concurrently

Day 16:
- Create:
  - `meal_logs`
  - `meal_log_items`

Day 17:
- Add:
  - `POST /meals/calculate`
- Input should support multiple ingredients

Day 18:
- Implement concurrent ingredient lookup
- Use goroutines and channels or `errgroup`
- Each ingredient should be matched independently
- Combine results safely

This is a central Go feature of the project.

Day 19:
- Build ambiguity detection
- Example:
  - `chicken` should return options, not a blind guess

Day 20:
- Add:
  - `POST /meals`
  - `GET /meals/today`

Day 21:
- Add:
  - `GET /summary/daily?date=YYYY-MM-DD`

Day 22:
- Build React meal logger and dashboard
- Show calories, protein, carbs, fat, and fiber

---

## Week 4: July 2–8 — AI Assistant Layer

Goal:
- User types natural language
- LLM extracts structured data
- Go calculates nutrition
- React shows result

Day 23:
- Design structured AI output schema:

```json
{
  "items": [
    {
      "name": "string",
      "quantity": "number",
      "unit": "string",
      "preparation": "string | null",
      "brand": "string | null"
    }
  ]
}
```

Day 24:
- Add:
  - `POST /ai/parse-meal`
- LLM returns structured JSON only

Day 25:
- Add:
  - `POST /chat/meal`
- Flow:
  1. User text
  2. LLM parses
  3. Go matches foods
  4. Go calculates nutrition
  5. Response returned

Day 26:
- Add clarification flow
- If Go returns ambiguity, ask user a follow-up

Day 27:
- Add coaching response
- Compare totals against user targets

Day 28:
- Build React chat UI
- Show chat messages plus nutrition cards

Day 29:
- Add Go unit tests for:
  - unit conversion
  - nutrition calculation
  - food matching
  - ambiguity detection
  - concurrent lookup aggregation

---

## Week 5: July 9–12 — MVP Polish

Goal:
- Project is demoable and resume-ready

Day 30:
- Add loading states, error states, and empty states

Day 31:
- Write README
- Include architecture diagram
- Explain why Go owns the nutrition engine

Day 32:
- Add demo dataset and screenshots
- Add example prompts:
  - `I had 2 homemade smash burgers and a Quest shake`
  - `230g chicken thighs and 150g cooked rice`
  - `Greek yogurt with whey and grapes`

Day 33:
- Write resume bullets

Resume bullets:
- Built a full-stack AI nutrition assistant (React/TypeScript, Go, PostgreSQL) that parses natural-language meal logs (e.g. "230g chicken thighs and a Quest shake") into calculated calories, protein, carbs, fat, and fiber, with a chat UI and structured logging form sharing one backend nutrition engine.
- Designed a Go nutrition engine with a repository/service architecture, typed domain models, unit conversion (weight and count-based foods via a stored grams-per-unit reference), and Postgres trigram (pg_trgm) fuzzy food matching with similarity-gap ambiguity detection.
- Implemented concurrent multi-ingredient resolution in Go using goroutines and errgroup, resolving each ingredient in a meal independently while preserving result order; covered with unit and race-detector tests (go test -race).
- Integrated the Anthropic Claude API using tool-use with an enum-constrained JSON schema to convert free-form meal text into validated structured requests, keeping all nutrition math deterministic and outside the LLM's responsibility.
- Implemented meal logging with transactional saves, daily nutrition summaries computed from live (non-snapshotted) food data, and a clarification flow that surfaces ambiguous or unmatched ingredients to the user instead of guessing.

---

## Post-MVP Roadmap

Phase 2: Docker
- Dockerize backend
- Dockerize frontend
- Manage external food API credentials via environment variables / secrets

Phase 3: MCP
- Expose existing Go backend through MCP tools:
  - `search_food`
  - `get_food_nutrition`
  - `calculate_meal`
  - `save_meal_log`
  - `get_daily_summary`

Phase 4: iPhone App
- Build with React Native + Expo
- Reuse same Go backend and AI parser

---

---

# Daily Coaching Template for Claude

When I start a session, ask me:

1. What day/task are we working on?
2. What have you already completed?
3. What is currently failing or confusing?

Then respond with:

1. The goal for today.
2. The concepts I need to understand.
3. A small implementation plan.
4. The first step I should do myself.

At the end of each session, help me write:

1. What I completed.
2. What I learned.
3. What remains broken or unfinished.
4. The next task.

---

# Quality Bar

This project should demonstrate:

- Clean full-stack architecture
- Practical AI integration
- Real backend logic, not just prompt engineering
- Typed frontend development
- Go API design
- PostgreSQL schema design
- Testing of deterministic nutrition calculations
- Human-in-the-loop AI workflows
- Resume-ready documentation

Do not let the project become a shallow AI wrapper.

The strongest version of this project is one where the LLM helps with natural language understanding, while the backend owns correctness.
