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

// brandHardThreshold/brandAmbiguousLow mirror foods/repository.go's
// constants of the same meaning — duplicated here rather than imported
// because they govern a decision (accept/reject/verify) that belongs to the
// matcher, not the repository. Keep these in sync with repository.go if
// tuning.
const brandHardThreshold = 0.6
const brandAmbiguousLow = 0.4

// searcher is the subset of foods.Service the matcher depends on. Defined
// here (not in foods) so tests can supply a fake without touching the DB.
type searcher interface {
	Resolve(ctx context.Context, query string) ([]foods.ScoredFood, error)
	ResolveBranded(ctx context.Context, brand, productName string) ([]foods.ScoredFood, error)
	FetchExternal(ctx context.Context, query, brand string) *nutrition.Food
}

// brandVerifier is the subset of ai.Client the matcher depends on for
// resolving ambiguous brand matches. Defined here so tests can fake it.
type brandVerifier interface {
	VerifyBrandMatch(ctx context.Context, queryBrand, queryProduct, candidateBrand, candidateName string) (bool, error)
}

type Matcher struct {
	foodService searcher
	verifier    brandVerifier
}

func NewMatcher(foodService *foods.Service, verifier brandVerifier) *Matcher {
	return &Matcher{foodService: foodService, verifier: verifier}
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

// searchQuery builds the food name used for the brand-agnostic path (no
// brand on the item) and for external/LLM lookups. Preparation is folded in
// since raw vs. cooked nutrition profiles differ meaningfully; see the
// brandedQuery doc comment for why brand is handled separately rather than
// folded into this same string.
func searchQuery(item ItemRequest) string {
	query := item.FoodName
	if item.Preparation != nil && *item.Preparation != "" {
		query = *item.Preparation + " " + query
	}
	return query
}

func brandOf(item ItemRequest) string {
	if item.Brand == nil {
		return ""
	}
	return *item.Brand
}

func (m *Matcher) resolveItem(ctx context.Context, item ItemRequest) ItemResult {
	result := ItemResult{
		FoodName: item.FoodName,
		Brand:    item.Brand,
		Quantity: item.Quantity,
		Unit:     item.Unit,
	}

	brand := brandOf(item)
	if brand != "" {
		return m.resolveBrandedItem(ctx, item, brand, result)
	}

	query := searchQuery(item)

	candidates, err := m.foodService.Resolve(ctx, query)
	if err != nil {
		result.Error = "lookup failed: " + err.Error()
		return result
	}

	if len(candidates) == 0 {
		if externalFood := m.foodService.FetchExternal(ctx, query, ""); externalFood != nil {
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

	return finalizeMatch(result, candidates[0].Food, item)
}

// resolveBrandedItem implements the two-stage retrieval + verification flow
// for a query that specifies a brand:
//
//  1. Candidate generation: trigram similarity over "brand name" combined,
//     forgiving of typos, returns up to 10 ranked candidates.
//  2. Verification: brand similarity is checked as a hard filter, SEPARATE
//     from the combined score. "Protein shake" matching across brands is
//     expected and fine; a different brand passing this filter is not — see
//     the "David protein bar" / "Quest protein shake" false-match this
//     replaced (searchQuery previously folded brand into one fuzzy string,
//     so a wrong-brand-right-category food could still slip through at the
//     bare similarityThreshold).
//     - brand_sim >= brandHardThreshold: accept.
//     - brand_sim < brandAmbiguousLow: reject outright, no LLM call.
//     - otherwise (the ambiguous middle): ask the LLM verifier, since this
//       is exactly the misspelling/abbreviation zone trigram similarity
//       can't resolve on its own.
//
// Falls through to FetchExternal (with brand preserved) when nothing in our
// own database survives verification, same unconfirmed-draft-then-human-
// review contract as the unbranded path.
func (m *Matcher) resolveBrandedItem(ctx context.Context, item ItemRequest, brand string, result ItemResult) ItemResult {
	productName := searchQuery(item)

	candidates, err := m.foodService.ResolveBranded(ctx, brand, productName)
	if err != nil {
		result.Error = "lookup failed: " + err.Error()
		return result
	}

	var verified []foods.ScoredFood
	for _, c := range candidates {
		switch {
		case c.BrandSimilarity >= brandHardThreshold:
			verified = append(verified, c)
		case c.BrandSimilarity < brandAmbiguousLow:
			// Confidently a different brand — reject without spending an
			// LLM call on it.
		default:
			candidateBrand := ""
			if c.Food.Brand != nil {
				candidateBrand = *c.Food.Brand
			}
			same, err := m.verifier.VerifyBrandMatch(ctx, brand, productName, candidateBrand, c.Food.Name)
			if err == nil && same {
				verified = append(verified, c)
			}
			// A verifier error or a "no" verdict both fall through to
			// rejection — an unconfirmed API/LLM failure should never
			// silently promote a brand match we couldn't actually confirm.
		}
	}

	if len(verified) == 0 {
		if externalFood := m.foodService.FetchExternal(ctx, brand+" "+productName, brand); externalFood != nil {
			result.UnconfirmedFood = externalFood
			return result
		}
		result.Error = "no matching food found"
		return result
	}

	if isAmbiguous(verified) {
		result.Ambiguous = true
		for _, c := range verified {
			result.Candidates = append(result.Candidates, c.Food)
		}
		return result
	}

	return finalizeMatch(result, verified[0].Food, item)
}

func finalizeMatch(result ItemResult, food nutrition.Food, item ItemRequest) ItemResult {
	calculated, err := nutrition.CalculateNutrition(food, item.Quantity, item.Unit)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.MatchedFood = &food
	result.Nutrition = &calculated
	return result
}
