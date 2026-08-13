package foods

import (
	"context"
	"log"
	"strings"

	"github.com/dhruvv301292/nutrichat/internal/nutrition"
)

// ExternalLookup is the subset of a fooddata client's behavior the service
// depends on. Defined here (not in fooddata) so tests can supply a fake
// without hitting real external APIs, and so fooddata doesn't need to
// depend on this package — main.go adapts the concrete provider clients to
// this shape when wiring the service up.
type ExternalLookup interface {
	Lookup(ctx context.Context, query string) (*ExternalResult, error)
}

// ExternalResult mirrors fooddata.Result. Declared here rather than
// importing the fooddata package directly, so foods doesn't depend on the
// concrete provider clients — main.go wires concrete *fooddata.USDAClient /
// *fooddata.FatSecretClient in, since both already satisfy this shape.
type ExternalResult struct {
	Name            string
	Calories        float64
	Protein         float64
	Carbs           float64
	Fat             float64
	Fiber           float64
	Sodium          float64
	GramsPerServing *float64
	// Unit/UnitQuantity, when set, are used as-is instead of the
	// GramsPerServing-based inference below — the AI estimator picks these
	// itself (grams/100 vs count/1) rather than always defaulting to a
	// 100g basis.
	Unit         string
	UnitQuantity float64
}

type Service struct {
	repo      *Repository
	providers []ExternalLookup
}

func NewService(repo *Repository, providers ...ExternalLookup) *Service {
	return &Service{repo: repo, providers: providers}
}

func (s *Service) Search(ctx context.Context, query string) ([]nutrition.Food, error) {
	normalized := strings.TrimSpace(query)
	return s.repo.Search(ctx, normalized)
}

func (s *Service) List(ctx context.Context) ([]nutrition.Food, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (nutrition.Food, error) {
	return s.repo.GetByID(ctx, id)
}

// Create persists a user-confirmed food — the shared endpoint for both the
// LLM-estimate flow and the external-database-match flow. Nothing from
// outside this service (an external API, an LLM guess) reaches the foods
// table without a human confirming it first, since both sources have shown
// themselves to be unreliable (see CLAUDE.md's Food Lookup Flow and the
// "chicken wings" incident — USDA's top-ranked result for a plain search
// can be an outlier branded product's numbers).
func (s *Service) Create(ctx context.Context, f nutrition.Food) (nutrition.Food, error) {
	return s.repo.Insert(ctx, f)
}

func (s *Service) SearchWithScore(ctx context.Context, query string) ([]ScoredFood, error) {
	normalized := strings.TrimSpace(query)
	return s.repo.SearchWithScore(ctx, normalized)
}

// FetchExternal tries each configured external food database in order and
// returns the first hit as an unconfirmed nutrition.Food (ID 0, not
// persisted). Callers must show this to the user for review/edit and save
// it via Create only once confirmed — see the Create doc comment for why.
// brand, if non-empty, is stamped onto the returned draft so a subsequent
// Create persists it as a proper branded row rather than losing that
// context (the LLM/external provider only sees the combined query string,
// not a structured brand field).
func (s *Service) FetchExternal(ctx context.Context, query, brand string) *nutrition.Food {
	for _, provider := range s.providers {
		result, err := provider.Lookup(ctx, query)
		if err != nil {
			log.Printf("fooddata provider lookup failed for %q: %v", query, err)
			continue
		}
		if result == nil {
			continue
		}

		food := nutrition.Food{
			Name:     result.Name,
			Calories: result.Calories,
			Protein:  result.Protein,
			Carbs:    result.Carbs,
			Fat:      result.Fat,
			Fiber:    result.Fiber,
			Sodium:   result.Sodium,
		}
		if brand != "" {
			food.Brand = &brand
		}
		switch {
		case result.Unit != "":
			food.Unit = result.Unit
			food.UnitQuantity = result.UnitQuantity
		case result.GramsPerServing != nil:
			food.Unit = "count"
			food.UnitQuantity = 1
			food.GramsPerUnit = result.GramsPerServing
		default:
			food.Unit = "grams"
			food.UnitQuantity = 100
		}
		return &food
	}
	return nil
}

// Resolve is the read-only lookup path for an already-logged-food query:
// Postgres only. Unlike the old ResolveOrFetch, this never reaches out to
// external providers or writes anything — matcher.resolveItem calls
// FetchExternal separately when this comes up empty, so it can surface the
// result as an unconfirmed draft instead of silently trusting it.
func (s *Service) Resolve(ctx context.Context, query string) ([]ScoredFood, error) {
	return s.SearchWithScore(ctx, query)
}

// ResolveBranded is the read-only two-stage lookup for a query that
// specifies a brand: candidate generation via combined brand+name trigram
// similarity, with each candidate's brand-only similarity carried alongside
// so the caller (Matcher) can apply the hard brand filter / ambiguous-zone
// LLM-verification logic — see repository.SearchBranded's doc comment for
// why that decision doesn't belong here.
func (s *Service) ResolveBranded(ctx context.Context, brand, productName string) ([]ScoredFood, error) {
	return s.repo.SearchBranded(ctx, strings.TrimSpace(brand), strings.TrimSpace(productName))
}
