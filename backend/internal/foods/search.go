package foods

import (
	"github.com/dhruvv301292/nutrichat/internal/nutrition"
	"strings"
)

func SearchFoods(query string) []nutrition.Food {
	foundFoods := []nutrition.Food{}
	for _, food := range SeedFoods {
		if strings.Contains(strings.ToLower(food.Name), strings.ToLower(query)) {
			foundFoods = append(foundFoods, food)
		}
	}
	return foundFoods
}
