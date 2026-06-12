package nutrition

type Food struct {
	Name     string  `json:"name"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
	Sodium   float64 `json:"sodium"`
}

type Nutrition struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
	Sodium   float64 `json:"sodium"`
}

func (n *Nutrition) Add(food Food, quantityInGrams float64) {
	n.Calories += (food.Calories/100) * quantityInGrams
	n.Protein += (food.Protein/100) * quantityInGrams
	n.Carbs += (food.Carbs/100) * quantityInGrams
	n.Fat += (food.Fat/100) * quantityInGrams
	n.Fiber += (food.Fiber/100) * quantityInGrams
	n.Sodium += (food.Sodium/100) * quantityInGrams
}