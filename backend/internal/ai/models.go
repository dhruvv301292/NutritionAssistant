package ai

// ParsedItem mirrors the structured output schema from CLAUDE.md's Day 22
// design: the LLM extracts these fields from free text and nothing else —
// no nutrition values, no matching against the food catalog. That work
// belongs to the Go nutrition engine.
type ParsedItem struct {
	Name        string  `json:"name"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	Preparation *string `json:"preparation"`
	Brand       *string `json:"brand"`
}

type ParsedMeal struct {
	Items []ParsedItem `json:"items"`
}

// NutritionEstimate is a last-resort, user-editable guess at a food's
// macros when it's in neither our own database nor any external food API.
// Per CLAUDE.md's Key Architecture Rule, this is explicitly NOT trusted as
// a calculation result — the client must show it as a draft the user can
// correct before it's saved to the foods table.
type NutritionEstimate struct {
	Name         string  `json:"name"`
	Calories     float64 `json:"calories"`
	Protein      float64 `json:"protein"`
	Carbs        float64 `json:"carbs"`
	Fat          float64 `json:"fat"`
	Fiber        float64 `json:"fiber"`
	Sodium       float64 `json:"sodium"`
	Unit         string  `json:"unit"`
	UnitQuantity float64 `json:"unitquantity"`
}
