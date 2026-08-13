package meals

import (
	"time"

	"github.com/dhruvv301292/nutrichat/internal/nutrition"
)

type ItemRequest struct {
	FoodName    string  `json:"food_name"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	Preparation *string `json:"preparation,omitempty"`
	Brand       *string `json:"brand,omitempty"`
}

type CalculateRequest struct {
	Items []ItemRequest `json:"items"`
}

type ItemResult struct {
	FoodName        string               `json:"food_name"`
	Quantity        float64              `json:"quantity"`
	Unit            string               `json:"unit"`
	MatchedFood     *nutrition.Food      `json:"matched_food,omitempty"`
	Nutrition       *nutrition.Nutrition `json:"nutrition,omitempty"`
	Ambiguous       bool                 `json:"ambiguous"`
	Candidates      []nutrition.Food     `json:"candidates,omitempty"`
	Error           string               `json:"error,omitempty"`
	UnconfirmedFood *nutrition.Food      `json:"unconfirmed_food,omitempty"`
}

type CalculateResponse struct {
	Items []ItemResult      `json:"items"`
	Total nutrition.Nutrition `json:"total"`
}

type Log struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`
	LoggedAt time.Time `json:"logged_at"`
	Slot     string    `json:"slot"`
	Items    []LogItem `json:"items"`
}

type LogItem struct {
	ID        int                  `json:"id"`
	FoodID    int                  `json:"food_id"`
	Food      *nutrition.Food      `json:"food,omitempty"`
	Quantity  float64              `json:"quantity"`
	Unit      string               `json:"unit"`
	Nutrition *nutrition.Nutrition `json:"nutrition,omitempty"`
}

type SaveRequest struct {
	Slot  string        `json:"slot"`
	Items []ItemRequest `json:"items"`
}

var ValidSlots = map[string]bool{
	"breakfast": true,
	"lunch":     true,
	"dinner":    true,
	"snack":     true,
}

type DailySummary struct {
	Date  string            `json:"date"`
	Total nutrition.Nutrition `json:"total"`
	Meals []Log             `json:"meals"`
}
