package meals

import (
	"context"
	"errors"
	"testing"

	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/dhruvv301292/nutrichat/internal/nutrition"
)

// fakeSearcher lets tests control exactly what candidates (and external
// fallback results) come back for a given query, without touching Postgres
// or a real external API.
type fakeSearcher struct {
	byQuery         map[string][]foods.ScoredFood
	externalByQuery map[string]*nutrition.Food
	err             error
}

func (f *fakeSearcher) Resolve(_ context.Context, query string) ([]foods.ScoredFood, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byQuery[query], nil
}

func (f *fakeSearcher) FetchExternal(_ context.Context, query string) *nutrition.Food {
	return f.externalByQuery[query]
}

func chickenThigh() nutrition.Food {
	return nutrition.Food{ID: 1, Name: "Chicken Thigh, Raw", Unit: "grams", UnitQuantity: 112, Calories: 145}
}

func chickenBreast() nutrition.Food {
	return nutrition.Food{ID: 2, Name: "Chicken Breast, Raw", Unit: "grams", UnitQuantity: 100, Calories: 120}
}

func TestIsAmbiguous_SingleCandidateNeverAmbiguous(t *testing.T) {
	if isAmbiguous([]foods.ScoredFood{{Food: chickenThigh(), Similarity: 0.9}}) {
		t.Error("expected single candidate to not be ambiguous")
	}
}

func TestIsAmbiguous_ClearLeaderNotAmbiguous(t *testing.T) {
	candidates := []foods.ScoredFood{
		{Food: chickenThigh(), Similarity: 0.9},
		{Food: chickenBreast(), Similarity: 0.3},
	}
	if isAmbiguous(candidates) {
		t.Error("expected a clear score gap to not be ambiguous")
	}
}

func TestIsAmbiguous_CloseScoresAreAmbiguous(t *testing.T) {
	candidates := []foods.ScoredFood{
		{Food: chickenThigh(), Similarity: 0.42},
		{Food: chickenBreast(), Similarity: 0.40},
	}
	if !isAmbiguous(candidates) {
		t.Error("expected a close score gap to be ambiguous")
	}
}

func TestMatcher_ResolveItems_PreservesInputOrder(t *testing.T) {
	fake := &fakeSearcher{byQuery: map[string][]foods.ScoredFood{
		"egg":     {{Food: nutrition.Food{ID: 3, Name: "Egg", Unit: "count", UnitQuantity: 1, Calories: 80}, Similarity: 0.9}},
		"chicken": {{Food: chickenThigh(), Similarity: 0.9}},
		"rice":    {}, // no match on purpose
	}}
	matcher := &Matcher{foodService: fake}

	items := []ItemRequest{
		{FoodName: "chicken", Quantity: 230, Unit: "grams"},
		{FoodName: "rice", Quantity: 150, Unit: "grams"},
		{FoodName: "egg", Quantity: 2, Unit: "count"},
	}

	results := matcher.ResolveItems(context.Background(), items)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].FoodName != "chicken" || results[0].MatchedFood == nil {
		t.Errorf("result[0]: expected matched chicken, got %+v", results[0])
	}
	if results[1].FoodName != "rice" || results[1].Error == "" {
		t.Errorf("result[1]: expected unresolved rice, got %+v", results[1])
	}
	if results[2].FoodName != "egg" || results[2].MatchedFood == nil {
		t.Errorf("result[2]: expected matched egg, got %+v", results[2])
	}
}

func TestMatcher_ResolveItems_AmbiguousItemReportsCandidates(t *testing.T) {
	fake := &fakeSearcher{byQuery: map[string][]foods.ScoredFood{
		"chicken": {
			{Food: chickenThigh(), Similarity: 0.42},
			{Food: chickenBreast(), Similarity: 0.40},
		},
	}}
	matcher := &Matcher{foodService: fake}

	results := matcher.ResolveItems(context.Background(), []ItemRequest{
		{FoodName: "chicken", Quantity: 100, Unit: "grams"},
	})

	if !results[0].Ambiguous {
		t.Fatal("expected item to be reported ambiguous")
	}
	if len(results[0].Candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(results[0].Candidates))
	}
	if results[0].MatchedFood != nil {
		t.Error("expected no matched food when ambiguous")
	}
}

func TestMatcher_ResolveItems_LookupErrorSurfacedPerItem(t *testing.T) {
	fake := &fakeSearcher{err: errors.New("connection refused")}
	matcher := &Matcher{foodService: fake}

	results := matcher.ResolveItems(context.Background(), []ItemRequest{
		{FoodName: "chicken", Quantity: 100, Unit: "grams"},
	})

	if results[0].Error == "" {
		t.Error("expected lookup error to be surfaced on the item")
	}
}

func TestMatcher_ResolveItems_ExternalMatchIsUnconfirmedNotAutoTrusted(t *testing.T) {
	externalFood := &nutrition.Food{Name: "CHICKEN WINGS", Unit: "grams", UnitQuantity: 100, Calories: 283}
	fake := &fakeSearcher{
		byQuery:         map[string][]foods.ScoredFood{}, // nothing in "our" DB
		externalByQuery: map[string]*nutrition.Food{"chicken wings": externalFood},
	}
	matcher := &Matcher{foodService: fake}

	results := matcher.ResolveItems(context.Background(), []ItemRequest{
		{FoodName: "chicken wings", Quantity: 100, Unit: "grams"},
	})

	if results[0].UnconfirmedFood == nil {
		t.Fatal("expected an unconfirmed food from the external provider")
	}
	if results[0].UnconfirmedFood.Name != "CHICKEN WINGS" {
		t.Errorf("unexpected unconfirmed food: %+v", results[0].UnconfirmedFood)
	}
	// The whole point: an external hit must never be auto-trusted as a
	// match — no calculated nutrition, no MatchedFood, until a human
	// confirms it via the Create endpoint.
	if results[0].MatchedFood != nil {
		t.Error("external match should not be treated as a confirmed MatchedFood")
	}
	if results[0].Nutrition != nil {
		t.Error("external match should not have calculated nutrition before confirmation")
	}
	if results[0].Error != "" {
		t.Errorf("expected no error for an external match awaiting confirmation, got %q", results[0].Error)
	}
}

func TestMatcher_ResolveItems_NoMatchAnywhereReportsError(t *testing.T) {
	fake := &fakeSearcher{byQuery: map[string][]foods.ScoredFood{}}
	matcher := &Matcher{foodService: fake}

	results := matcher.ResolveItems(context.Background(), []ItemRequest{
		{FoodName: "nonexistent food", Quantity: 100, Unit: "grams"},
	})

	if results[0].Error == "" {
		t.Error("expected an error when nothing matches locally or externally")
	}
	if results[0].UnconfirmedFood != nil {
		t.Error("expected no unconfirmed food when no external provider had a match")
	}
}
