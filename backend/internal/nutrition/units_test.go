package nutrition

import "testing"

func gramsPerUnitPtr(v float64) *float64 { return &v }

func TestConvertToGrams_GramsToGrams(t *testing.T) {
	food := Food{Unit: "grams", UnitQuantity: 112}
	got, err := ConvertToGrams(230, "grams", food)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 230 {
		t.Errorf("got %v, want 230", got)
	}
}

func TestConvertToGrams_OuncesToGrams(t *testing.T) {
	food := Food{Unit: "grams", UnitQuantity: 112}
	got, err := ConvertToGrams(8, "ounces", food)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := 8 * gramsPerOunce
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConvertToGrams_UnsupportedUnitForGramFood(t *testing.T) {
	food := Food{Unit: "grams", UnitQuantity: 112}
	_, err := ConvertToGrams(1, "cups", food)
	if err == nil {
		t.Fatal("expected error for unsupported unit, got nil")
	}
}

func TestConvertToGrams_CountFoodNativeUnit(t *testing.T) {
	food := Food{Unit: "count", UnitQuantity: 1, GramsPerUnit: gramsPerUnitPtr(64)}
	got, err := ConvertToGrams(2, "count", food)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("got %v, want 2", got)
	}
}

func TestConvertToGrams_CountFoodFromGrams(t *testing.T) {
	food := Food{Unit: "count", UnitQuantity: 1, GramsPerUnit: gramsPerUnitPtr(64)}
	got, err := ConvertToGrams(100, "grams", food)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := 100.0 / 64.0
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConvertToGrams_CountFoodFromOunces(t *testing.T) {
	food := Food{Unit: "count", UnitQuantity: 1, GramsPerUnit: gramsPerUnitPtr(64)}
	got, err := ConvertToGrams(1, "ounces", food)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := gramsPerOunce / 64.0
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConvertToGrams_CountFoodWithoutGramsPerUnit_RequiresExactMatch(t *testing.T) {
	food := Food{Unit: "count", UnitQuantity: 1}
	if _, err := ConvertToGrams(1, "count", food); err != nil {
		t.Errorf("expected native unit to succeed, got error: %v", err)
	}
	if _, err := ConvertToGrams(1, "grams", food); err == nil {
		t.Error("expected error converting grams for a food with no GramsPerUnit, got nil")
	}
}

func TestConvertToGrams_MismatchedUnitErrors(t *testing.T) {
	food := Food{Unit: "count", UnitQuantity: 1}
	if _, err := ConvertToGrams(1, "bottle", food); err == nil {
		t.Error("expected error for mismatched unit, got nil")
	}
}
