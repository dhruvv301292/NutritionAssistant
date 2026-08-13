package goals

// ValidTrackedMacros are the macros a user can individually show/hide.
// Calories are always shown everywhere and aren't part of this set.
var ValidTrackedMacros = map[string]bool{
	"protein": true,
	"carbs":   true,
	"fat":     true,
	"fiber":   true,
	"sodium":  true,
}

// DefaultTrackedMacros mirrors what the UI has always shown before this
// preference existed, so upgrading users see no change until they opt in
// (sodium is new and starts hidden).
var DefaultTrackedMacros = []string{"protein", "carbs", "fat", "fiber"}

type Goals struct {
	UserID        int      `json:"user_id"`
	CalorieGoal   int      `json:"calorie_goal"`
	ProteinGoal   int      `json:"protein_goal"`
	CarbGoal      int      `json:"carb_goal"`
	FatGoal       int      `json:"fat_goal"`
	FiberGoal     int      `json:"fiber_goal"`
	SodiumGoal    int      `json:"sodium_goal"`
	TrackedMacros []string `json:"tracked_macros"`
}

// Defaults mirror the mockup's initial targets so a first-time user sees
// sensible numbers before ever setting their own. SodiumGoal defaults to
// the AHA's stricter recommended limit.
func Defaults(userID int) Goals {
	return Goals{
		UserID:        userID,
		CalorieGoal:   2200,
		ProteinGoal:   140,
		CarbGoal:      250,
		FatGoal:       70,
		FiberGoal:     30,
		SodiumGoal:    1800,
		TrackedMacros: DefaultTrackedMacros,
	}
}
