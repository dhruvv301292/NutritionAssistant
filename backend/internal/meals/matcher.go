package meals

import (
	"context"

	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/dhruvv301292/nutrichat/internal/nutrition"
	"golang.org/x/sync/errgroup"
)

// ambiguityGap is the minimum similarity-score lead the top candidate needs
// over the runner-up before we're willing to auto-pick it. Below this, we
// report the item as ambiguous instead of guessing.
const ambiguityGap = 0.05

// searcher is the subset of foods.Service the matcher depends on. Defined
// here (not in foods) so tests can supply a fake without touching the DB.
type searcher interface {
	Resolve(ctx context.Context, query string) ([]foods.ScoredFood, error)
	FetchExternal(ctx context.Context, query string) *nutrition.Food
}

type Matcher struct {
	foodService searcher
}

func NewMatcher(foodService *foods.Service) *Matcher {
	return &Matcher{foodService: foodService}
}

// ResolveItems matches every requested item against the food catalog
// concurrently — one goroutine per item — and returns results in the same
// order as the input, regardless of which goroutine finishes first.
func (m *Matcher) ResolveItems(ctx context.Context, items []ItemRequest) []ItemResult {
	results := make([]ItemResult, len(items))

	g, ctx := errgroup.WithContext(ctx)
	for i, item := range items {
		i, item := i, item // capture loop variables per-goroutine
		g.Go(func() error {
			results[i] = m.resolveItem(ctx, item)
			return nil
		})
	}
	// Errors are captured per-item in ItemResult.Error rather than failing
	// the whole batch, so g.Wait()'s error is always nil here.
	_ = g.Wait()

	return results
}

// isAmbiguous reports whether the top-scored candidate isn't clearly ahead
// of the runner-up. A single candidate is never ambiguous.
func isAmbiguous(candidates []foods.ScoredFood) bool {
	return len(candidates) > 1 && candidates[0].Similarity-candidates[1].Similarity < ambiguityGap
}

func (m *Matcher) resolveItem(ctx context.Context, item ItemRequest) ItemResult {
	result := ItemResult{
		FoodName: item.FoodName,
		Quantity: item.Quantity,
		Unit:     item.Unit,
	}

	candidates, err := m.foodService.Resolve(ctx, item.FoodName)
	if err != nil {
		result.Error = "lookup failed: " + err.Error()
		return result
	}

	if len(candidates) == 0 {
		// Nothing in our own database — try external food databases, but
		// surface any hit as an unconfirmed draft rather than trusting it
		// automatically. External data (USDA in particular) has shown up
		// with wrong or missing values for plain-language queries, so a
		// human reviews/edits before it's saved to the foods table.
		if externalFood := m.foodService.FetchExternal(ctx, item.FoodName); externalFood != nil {
			result.UnconfirmedFood = externalFood
			return result
		}
		result.Error = "no matching food found"
		return result
	}

	if isAmbiguous(candidates) {
		result.Ambiguous = true
		for _, c := range candidates {
			result.Candidates = append(result.Candidates, c.Food)
		}
		return result
	}

	best := candidates[0].Food
	calculated, err := nutrition.CalculateNutrition(best, item.Quantity, item.Unit)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.MatchedFood = &best
	result.Nutrition = &calculated
	return result
}
