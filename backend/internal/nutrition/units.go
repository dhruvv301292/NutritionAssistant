package nutrition

import "errors"

const gramsPerOunce = 28.3495

// ConvertToGrams converts a user-supplied quantity/unit into whatever unit
// food.UnitQuantity is expressed in (grams for weight-based foods, a native
// count for foods with GramsPerUnit set, e.g. eggs).
func ConvertToGrams(quantity float64, unit string, food Food) (float64, error) {
	if food.Unit == "grams" {
		switch unit {
		case "grams":
			return quantity, nil
		case "ounces":
			return quantity * gramsPerOunce, nil
		default:
			return 0, errors.New("unsupported unit for gram-based food")
		}
	}

	if food.GramsPerUnit != nil {
		switch unit {
		case food.Unit:
			return quantity, nil
		case "grams":
			return quantity / *food.GramsPerUnit, nil
		case "ounces":
			return quantity * gramsPerOunce / *food.GramsPerUnit, nil
		default:
			return 0, errors.New("unsupported unit for this food")
		}
	}

	if unit != food.Unit {
		return 0, errors.New("unit does not match food's native unit")
	}
	return quantity, nil
}
