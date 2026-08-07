package goals

type Goals struct {
	UserID      int `json:"user_id"`
	CalorieGoal int `json:"calorie_goal"`
	ProteinGoal int `json:"protein_goal"`
	CarbGoal    int `json:"carb_goal"`
	FatGoal     int `json:"fat_goal"`
	FiberGoal   int `json:"fiber_goal"`
}

// Defaults mirror the mockup's initial targets so a first-time user sees
// sensible numbers before ever setting their own.
func Defaults(userID int) Goals {
	return Goals{
		UserID:      userID,
		CalorieGoal: 2200,
		ProteinGoal: 140,
		CarbGoal:    250,
		FatGoal:     70,
		FiberGoal:   30,
	}
}
