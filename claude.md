# Claude Coaching Guide: AI Nutrition Assistant Website

## Project Overview

You are coaching me while I build an AI Nutrition Assistant as a resume-quality full-stack project.

The project will start as a website and later support an iPhone app using the same backend.

Core goals:

1. Learn React + TypeScript by building the web frontend.
2. Learn Go by building the backend API and nutrition engine.
3. Learn PostgreSQL through real schema design and querying.
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

## How You Should Coach Me

Act as my instructor, tutor, and code reviewer.

For each task:

1. Explain the concept before giving code.
2. Ask me to attempt the implementation first when reasonable.
3. Review my code critically.
4. Explain mistakes in plain language.
5. Encourage clean architecture, not shortcuts.
6. Keep the project resume-quality.
7. Help me understand why each design choice matters.

When giving code:

- Prefer small, focused snippets over huge copy-paste files.
- Explain where each file goes.
- Explain what each function or component is responsible for.
- Ask me to run tests or manually verify behavior.
- Point out edge cases.

Do not let me rely on the LLM for nutrition math. The backend should calculate nutrition deterministically.

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
- Build homepage
- Build static `FoodSearch` and `NutritionCard` components

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
- Create local Postgres database
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
  5. Ranked candidates

Day 13:
- Build unit conversion v1:
  - grams
  - ounces
  - serving
- Function:
  - `ConvertToGrams(quantity float64, unit string, food Food)`

Day 14:
- Improve React food search UI
- Show serving size and nutrition per 100g

---

## Week 3: June 25–July 1 — Meal Logging + Go Concurrency

Goal:
- User can log a meal
- Backend calculates totals
- Go resolves multiple ingredients concurrently

Day 15:
- Create:
  - `meal_logs`
  - `meal_log_items`

Day 16:
- Add:
  - `POST /meals/calculate`
- Input should support multiple ingredients

Day 17:
- Implement concurrent ingredient lookup
- Use goroutines and channels or `errgroup`
- Each ingredient should be matched independently
- Combine results safely

This is a central Go feature of the project.

Day 18:
- Build ambiguity detection
- Example:
  - `chicken` should return options, not a blind guess

Day 19:
- Add:
  - `POST /meals`
  - `GET /meals/today`

Day 20:
- Add:
  - `GET /summary/daily?date=YYYY-MM-DD`

Day 21:
- Build React meal logger and dashboard
- Show calories, protein, carbs, fat, and fiber

---

## Week 4: July 2–8 — AI Assistant Layer

Goal:
- User types natural language
- LLM extracts structured data
- Go calculates nutrition
- React shows result

Day 22:
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

Day 23:
- Add:
  - `POST /ai/parse-meal`
- LLM returns structured JSON only

Day 24:
- Add:
  - `POST /chat/meal`
- Flow:
  1. User text
  2. LLM parses
  3. Go matches foods
  4. Go calculates nutrition
  5. Response returned

Day 25:
- Add clarification flow
- If Go returns ambiguity, ask user a follow-up

Day 26:
- Add coaching response
- Compare totals against user targets

Day 27:
- Build React chat UI
- Show chat messages plus nutrition cards

Day 28:
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

Day 29:
- Add loading states, error states, and empty states

Day 30:
- Write README
- Include architecture diagram
- Explain why Go owns the nutrition engine

Day 31:
- Add demo dataset and screenshots
- Add example prompts:
  - `I had 2 homemade smash burgers and a Quest shake`
  - `230g chicken thighs and 150g cooked rice`
  - `Greek yogurt with whey and grapes`

Day 32:
- Write resume bullets

Suggested resume bullets:
- Built an AI-powered nutrition assistant using React, TypeScript, Go, and PostgreSQL to parse natural language meal logs and calculate calories, macros, and fiber.
- Designed a Go nutrition engine with typed domain models, unit conversion, deterministic food matching, ambiguity detection, and concurrent ingredient lookup.
- Integrated LLM structured outputs to convert free-form meal descriptions into validated backend requests while keeping nutrition calculations deterministic in Go.
- Implemented daily meal logging, nutrition summaries, and reusable APIs designed for future MCP and mobile app integration.

---

## Post-MVP Roadmap

Phase 2: Docker
- Dockerize backend
- Dockerize frontend
- Add docker-compose for Postgres

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
