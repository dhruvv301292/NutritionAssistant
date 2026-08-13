package fooddata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const usdaSearchURL = "https://api.nal.usda.gov/fdc/v1/foods/search"

// USDA nutrient numbers — stable across dataTypes, unlike nutrientId.
const (
	usdaNutrientEnergy   = "208"
	usdaNutrientProtein  = "203"
	usdaNutrientCarbs    = "205"
	usdaNutrientFat      = "204"
	usdaNutrientFiber    = "291"
	usdaNutrientSodium   = "307"
)

type USDAClient struct {
	apiKey string
	http   *http.Client
}

func NewUSDAClient(apiKey string) *USDAClient {
	return &USDAClient{apiKey: apiKey, http: &http.Client{}}
}

type usdaSearchResponse struct {
	Foods []usdaFood `json:"foods"`
}

type usdaFood struct {
	FDCID         int               `json:"fdcId"`
	Description   string            `json:"description"`
	DataType      string            `json:"dataType"`
	ServingSize   float64           `json:"servingSize"`
	ServingSizeUnit string          `json:"servingSizeUnit"`
	FoodNutrients []usdaFoodNutrient `json:"foodNutrients"`
}

type usdaFoodNutrient struct {
	NutrientNumber string  `json:"nutrientNumber"`
	Value          float64 `json:"value"`
}

// usdaGenericDataTypes are USDA's own curated reference datasets (single
// ingredient, lab-analyzed) as opposed to Branded/Survey entries, which are
// individual products or as-eaten mixed dishes. Searching restricted to
// these first, and only falling back to an unrestricted search if that
// comes up empty, matters because plain-language queries like "chicken
// wings" rank Branded products first in USDA's default relevance search —
// the generic "Chicken, wing, meat and skin, raw" reference exists but
// doesn't appear anywhere in the top 10 unrestricted results, so filtering
// after the fact (rather than restricting the query itself) doesn't help.
const usdaGenericDataTypes = "Foundation,SR Legacy"

// Lookup searches USDA FoodData Central for query. It first searches
// restricted to generic/reference data (see usdaGenericDataTypes), since
// that's what a plain food name like "chicken wings" should resolve to by
// default; only if that comes up empty does it fall back to an
// unrestricted search, which may return a Branded product. USDA's
// Foundation/SR Legacy data is per-100g; Branded foods carry their own
// servingSize instead, so we treat the latter as GramsPerServing when
// present.
func (c *USDAClient) Lookup(ctx context.Context, query string) (*Result, error) {
	food, err := c.search(ctx, query, usdaGenericDataTypes)
	if err != nil {
		return nil, err
	}
	if food != nil {
		if result := toResult(food); result != nil {
			return result, nil
		}
		// Generic entry exists but is missing usable nutrition data (seen
		// in practice: a Foundation food with no Energy value) — fall
		// through to an unrestricted search rather than returning nothing.
	}

	food, err = c.search(ctx, query, "")
	if err != nil {
		return nil, err
	}
	if food == nil {
		return nil, nil
	}
	return toResult(food), nil
}

// cookedKeywords/rawKeywords classify a USDA food description by
// preparation state, matched case-insensitively against the description
// text. USDA's own wording varies ("cooked", "roasted", "boiled", etc.)
// rather than using a single consistent term.
var (
	cookedKeywords = []string{"cooked", "roasted", "boiled", "grilled", "baked", "braised", "steamed", "fried", "broiled", "poached"}
	rawKeywords    = []string{"raw"}
)

func describesCooked(description string) bool {
	lower := strings.ToLower(description)
	for _, kw := range cookedKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func describesRaw(description string) bool {
	lower := strings.ToLower(description)
	for _, kw := range rawKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// wantsCooked/wantsRaw report whether the caller's query itself asked for a
// specific preparation state (matcher.externalQuery prefixes the food name
// with the parsed preparation, e.g. "cooked chicken thigh").
func wantsCooked(query string) bool { return describesCooked(query) }
func wantsRaw(query string) bool    { return describesRaw(query) }

// toResult normalizes a usdaFood into a Result, or nil if it's missing
// calories entirely — USDA occasionally has reference entries with no
// Energy value populated, and a zero-calorie food is worse to insert than
// no match at all (silently wrong instead of visibly absent). Negative
// macro values (a rounding artifact of USDA's "by difference" carb
// calculation) are clamped to zero rather than propagated.
func toResult(food *usdaFood) *Result {
	result := &Result{Name: food.Description}
	hasCalories := false
	for _, n := range food.FoodNutrients {
		value := n.Value
		if value < 0 {
			value = 0
		}
		switch n.NutrientNumber {
		case usdaNutrientEnergy:
			result.Calories = value
			hasCalories = n.Value != 0
		case usdaNutrientProtein:
			result.Protein = value
		case usdaNutrientCarbs:
			result.Carbs = value
		case usdaNutrientFat:
			result.Fat = value
		case usdaNutrientFiber:
			result.Fiber = value
		case usdaNutrientSodium:
			result.Sodium = value
		}
	}
	if !hasCalories {
		return nil
	}

	if food.DataType == "Branded" && food.ServingSize > 0 && food.ServingSizeUnit == "g" {
		servingGrams := food.ServingSize
		result.GramsPerServing = &servingGrams
	}

	return result
}

// usdaCandidatePoolSize is how many top-ranked USDA results we fetch and
// rerank locally, rather than blindly trusting result #1. USDA's relevance
// ranking doesn't account for raw/cooked state at all, so when the caller's
// query specifies one (see matcher.externalQuery), the correct entry is
// often #2 or #3, not #1.
const usdaCandidatePoolSize = 5

// search runs a single USDA search request, optionally restricted to the
// given comma-separated dataType list, and returns the best-matching result
// (nil if there were none). "Best" means: if the query specifies a
// preparation state (raw/cooked), prefer the top-ranked candidate whose
// description agrees with that state over USDA's raw #1 ranking; otherwise
// just take #1.
func (c *USDAClient) search(ctx context.Context, query, dataTypes string) (*usdaFood, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", query)
	params.Set("pageSize", fmt.Sprintf("%d", usdaCandidatePoolSize))
	if dataTypes != "" {
		params.Set("dataType", dataTypes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usdaSearchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("usda request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usda request failed: status %d", resp.StatusCode)
	}

	var parsed usdaSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("usda response decode failed: %w", err)
	}
	if len(parsed.Foods) == 0 {
		return nil, nil
	}

	if wantsCooked(query) {
		for i := range parsed.Foods {
			if describesCooked(parsed.Foods[i].Description) {
				return &parsed.Foods[i], nil
			}
		}
	} else if wantsRaw(query) {
		for i := range parsed.Foods {
			if describesRaw(parsed.Foods[i].Description) {
				return &parsed.Foods[i], nil
			}
		}
	}
	return &parsed.Foods[0], nil
}
