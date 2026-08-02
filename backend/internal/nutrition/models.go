package nutrition

type Food struct {
	Name         string  `json:"name"`
	Calories     float64 `json:"calories"`
	Protein      float64 `json:"protein"`
	Carbs        float64 `json:"carbs"`
	Fat          float64 `json:"fat"`
	Fiber        float64 `json:"fiber"`
	Sodium       float64 `json:"sodium"`
	Unit         string  `json:"unit"`
	UnitQuantity float64 `json:"unitquantity"`
}

type Nutrition struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
	Sodium   float64 `json:"sodium"`
}

func (n *Nutrition) Add(food Food, quantityConsumed float64) {
	n.Calories += (food.Calories / food.UnitQuantity) * quantityConsumed
	n.Protein += (food.Protein / food.UnitQuantity) * quantityConsumed
	n.Carbs += (food.Carbs / food.UnitQuantity) * quantityConsumed
	n.Fat += (food.Fat / food.UnitQuantity) * quantityConsumed
	n.Fiber += (food.Fiber / food.UnitQuantity) * quantityConsumed
	n.Sodium += (food.Sodium / food.UnitQuantity) * quantityConsumed
}

type CalculateNutritionRequest struct {
    FoodName string  `json:"food_name"`
    Quantity float64 `json:"quantity"`
    Unit     string  `json:"unit"`
}
