package nutrition

import "testing"

func TestCalculateNutrition_ZeroQuantityErrors(t *testing.T) {
	food := Food{Unit: "grams", UnitQuantity: 100, Calories: 100}
	_, err := CalculateNutrition(food, 0, "grams")
	if err == nil {
		t.Fatal("expected error for zero quantity, got nil")
	}
}

func TestCalculateNutrition_ScalesLinearly(t *testing.T) {
	food := Food{
		Unit: "grams", UnitQuantity: 112,
		Calories: 145, Protein: 21, Carbs: 0, Fat: 7, Fiber: 0, Sodium: 175,
	}
	got, err := CalculateNutrition(food, 230, "grams")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantCalories := (145.0 / 112.0) * 230.0
	if diff := got.Calories - wantCalories; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("calories: got %v, want %v", got.Calories, wantCalories)
	}
}

func TestCalculateNutrition_ConvertsOuncesBeforeScaling(t *testing.T) {
	food := Food{Unit: "grams", UnitQuantity: 112, Calories: 145}
	got, err := CalculateNutrition(food, 8, "ounces")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantGrams := 8 * gramsPerOunce
	wantCalories := (145.0 / 112.0) * wantGrams
	if diff := got.Calories - wantCalories; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("calories: got %v, want %v", got.Calories, wantCalories)
	}
}

func TestCalculateNutrition_CountFoodNativeUnit(t *testing.T) {
	food := Food{Unit: "count", UnitQuantity: 1, Calories: 80}
	got, err := CalculateNutrition(food, 2, "count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Calories != 160 {
		t.Errorf("calories: got %v, want 160", got.Calories)
	}
}

func TestCalculateNutrition_InvalidUnitErrors(t *testing.T) {
	food := Food{Unit: "grams", UnitQuantity: 100, Calories: 100}
	_, err := CalculateNutrition(food, 100, "cups")
	if err == nil {
		t.Fatal("expected error for invalid unit, got nil")
	}
}
