package nutrition

import (
	"errors"
)

func CalculateNutrition(food Food, quantity float64, unit string) (Nutrition, error) {
	if quantity == 0 {
		return Nutrition{}, errors.New("quantity cannot be zero")
	}
	returnNutrition := Nutrition{}
	quantity, err := ConvertToGrams(quantity, unit, food)
	if err != nil {
		return returnNutrition, err
	}
	returnNutrition.Add(food, quantity)

	return returnNutrition, nil
}