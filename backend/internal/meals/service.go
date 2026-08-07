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

type Service struct {
	repo        *Repository
	matcher     *Matcher
	foodService *foods.Service
}

func NewService(repo *Repository, matcher *Matcher, foodService *foods.Service) *Service {
	return &Service{repo: repo, matcher: matcher, foodService: foodService}
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
func (s *Service) Save(ctx context.Context, userID int, items []ItemRequest) (Log, []ItemResult, error) {
	results := s.matcher.ResolveItems(ctx, items)

	logItems := make([]LogItem, 0, len(results))
	for i, r := range results {
		if r.Error != "" || r.Ambiguous || r.MatchedFood == nil {
			return Log{}, results, ErrUnresolvedItems
		}
		logItems = append(logItems, LogItem{
			FoodID:   r.MatchedFood.ID,
			Quantity: items[i].Quantity,
			Unit:     items[i].Unit,
		})
	}

	log, err := s.repo.Save(ctx, userID, logItems)
	if err != nil {
		return Log{}, nil, err
	}
	return log, nil, nil
}

func (s *Service) Today(ctx context.Context, userID int) ([]Log, error) {
	now := time.Now().UTC()
	start := now.Format("2006-01-02")
	end := now.AddDate(0, 0, 1).Format("2006-01-02")
	logs, err := s.repo.ListByUserAndDateRange(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}
	return s.hydrate(ctx, logs)
}

func (s *Service) DailySummary(ctx context.Context, userID int, date string) (DailySummary, error) {
	start, err := time.Parse("2006-01-02", date)
	if err != nil {
		return DailySummary{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", date)
	}
	end := start.AddDate(0, 0, 1)

	logs, err := s.repo.ListByUserAndDateRange(ctx, userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
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
