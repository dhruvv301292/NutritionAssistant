package nutrition

import (
	"errors"
)

func CalculateNutrition(food Food, quantity float64, unit string) (Nutrition, error) {
	if quantity == 0 {
		return Nutrition{}, errors.New("quantity cannot be zero")
	}
	if unit != food.Unit {
		return Nutrition{}, errors.New("food unit and passed unit must match")
	}
	returnNutrition := Nutrition{}
	returnNutrition.Add(food, quantity)

	return returnNutrition, nil
}