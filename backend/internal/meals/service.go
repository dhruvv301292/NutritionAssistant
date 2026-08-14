package meals

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/dhruvv301292/nutrichat/internal/nutrition"
)

var ErrUnresolvedItems = errors.New("one or more items could not be resolved")
var ErrInvalidSlot = errors.New("slot must be one of breakfast, lunch, dinner, snack")

// titleGenerator is the subset of ai.Client Save depends on for the fun
// meal-title feature. Defined here so tests can fake it, same pattern as
// searcher/brandVerifier in matcher.go.
type titleGenerator interface {
	GenerateMealTitle(ctx context.Context, ingredients []string) (string, error)
}

type Service struct {
	repo        *Repository
	matcher     *Matcher
	foodService *foods.Service
	titleGen    titleGenerator
}

func NewService(repo *Repository, matcher *Matcher, foodService *foods.Service, titleGen titleGenerator) *Service {
	return &Service{repo: repo, matcher: matcher, foodService: foodService, titleGen: titleGen}
}

func (s *Service) Calculate(ctx context.Context, items []ItemRequest) CalculateResponse {
	results := s.matcher.ResolveItems(ctx, items)

	total := nutrition.Nutrition{}
	for _, r := range results {
		if r.Nutrition != nil {
			total.Calories += r.Nutrition.Calories
			total.Protein += r.Nutrition.Protein
			total.Carbs += r.Nutrition.Carbs
			total.Fat += r.Nutrition.Fat
			total.Fiber += r.Nutrition.Fiber
			total.Sodium += r.Nutrition.Sodium
		}
	}

	return CalculateResponse{Items: results, Total: total}
}

// Save resolves every item and, only if every single one matched
// unambiguously and without error, persists the meal. Otherwise nothing is
// written and the caller gets back the same per-item diagnostics Calculate
// would have produced, so the client can show the user what needs fixing.
func (s *Service) Save(ctx context.Context, userID int, slot string, items []ItemRequest) (Log, []ItemResult, error) {
	if !ValidSlots[slot] {
		return Log{}, nil, ErrInvalidSlot
	}

	results := s.matcher.ResolveItems(ctx, items)

	logItems := make([]LogItem, 0, len(results))
	ingredients := make([]string, 0, len(results))
	for i, r := range results {
		if r.Error != "" || r.Ambiguous || r.MatchedFood == nil {
			return Log{}, results, ErrUnresolvedItems
		}
		logItems = append(logItems, LogItem{
			FoodID:   r.MatchedFood.ID,
			Quantity: items[i].Quantity,
			Unit:     items[i].Unit,
		})
		ingredients = append(ingredients, r.MatchedFood.Name)
	}

	// Title generation is best-effort cosmetic flair, not part of the
	// nutrition pipeline's correctness — a failure here (rate limit, API
	// hiccup) must never block saving the actual meal, so errors are
	// swallowed and the meal is just saved without a title.
	var title *string
	if s.titleGen != nil {
		if t, err := s.titleGen.GenerateMealTitle(ctx, ingredients); err == nil && t != "" {
			title = &t
		}
	}

	log, err := s.repo.Save(ctx, userID, slot, title, logItems)
	if err != nil {
		return Log{}, nil, err
	}
	return log, nil, nil
}

// tzOffsetMinutes follows JS Date.getTimezoneOffset() convention: minutes
// the local zone is BEHIND UTC (e.g. UTC-5 => +300). Subtracting it from a
// UTC instant converts to local wall-clock time, and adding it converts back.
func (s *Service) Today(ctx context.Context, userID int, tzOffsetMinutes int) ([]Log, error) {
	localNow := time.Now().UTC().Add(-time.Duration(tzOffsetMinutes) * time.Minute)
	localDate := localNow.Format("2006-01-02")
	logs, err := s.dailySummaryLogs(ctx, userID, localDate, tzOffsetMinutes)
	if err != nil {
		return nil, err
	}
	return s.hydrate(ctx, logs)
}

func (s *Service) dailySummaryLogs(ctx context.Context, userID int, date string, tzOffsetMinutes int) ([]Log, error) {
	localStart, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", date)
	}
	localEnd := localStart.AddDate(0, 0, 1)

	offset := time.Duration(tzOffsetMinutes) * time.Minute
	utcStart := localStart.Add(offset)
	utcEnd := localEnd.Add(offset)

	return s.repo.ListByUserAndDateRange(ctx, userID, utcStart.Format("2006-01-02T15:04:05"), utcEnd.Format("2006-01-02T15:04:05"))
}

func (s *Service) DailySummary(ctx context.Context, userID int, date string, tzOffsetMinutes int) (DailySummary, error) {
	logs, err := s.dailySummaryLogs(ctx, userID, date, tzOffsetMinutes)
	if err != nil {
		return DailySummary{}, err
	}
	logs, err = s.hydrate(ctx, logs)
	if err != nil {
		return DailySummary{}, err
	}

	total := nutrition.Nutrition{}
	for _, log := range logs {
		for _, item := range log.Items {
			if item.Nutrition == nil {
				continue
			}
			total.Calories += item.Nutrition.Calories
			total.Protein += item.Nutrition.Protein
			total.Carbs += item.Nutrition.Carbs
			total.Fat += item.Nutrition.Fat
			total.Fiber += item.Nutrition.Fiber
			total.Sodium += item.Nutrition.Sodium
		}
	}

	return DailySummary{Date: date, Total: total, Meals: logs}, nil
}

// hydrate looks up each item's food record (by current data, not a
// point-in-time snapshot — see CLAUDE.md's cache-aside notes) and computes
// its nutrition, so callers can display food names/macros without a second
// round trip.
func (s *Service) hydrate(ctx context.Context, logs []Log) ([]Log, error) {
	for i := range logs {
		for j := range logs[i].Items {
			item := &logs[i].Items[j]
			food, err := s.foodService.GetByID(ctx, item.FoodID)
			if err != nil {
				continue
			}
			item.Food = &food
			calculated, err := nutrition.CalculateNutrition(food, item.Quantity, item.Unit)
			if err != nil {
				continue
			}
			item.Nutrition = &calculated
		}
	}
	return logs, nil
}
