package fooddata

// Result is the normalized shape both external food databases map their
// responses into, so callers (foods.Service) don't need to know which
// provider a match came from. Values are always on a per-100g basis for
// weight-based foods; count-based/branded items with a labeled serving use
// GramsPerServing to convert, mirroring nutrition.Food's own GramsPerUnit.
type Result struct {
	Name           string
	Calories       float64
	Protein        float64
	Carbs          float64
	Fat            float64
	Fiber          float64
	Sodium         float64
	GramsPerServing *float64 // nil means values are already per 100g
}
