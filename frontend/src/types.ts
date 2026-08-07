export type Food = {
  id: number,
  name: string,
  calories: number,
  protein: number,
  carbs: number,
  fat: number,
  fiber: number,
  sodium: number,
  unit: string,
  unitquantity: number,
  grams_per_unit?: number,
}

export type Nutrition = {
  calories: number,
  protein: number,
  carbs: number,
  fat: number,
  fiber: number,
  sodium: number,
}

export type ItemRequest = {
  food_name: string,
  quantity: number,
  unit: string,
}

export type ItemResult = {
  food_name: string,
  quantity: number,
  unit: string,
  matched_food?: Food,
  nutrition?: Nutrition,
  ambiguous: boolean,
  candidates?: Food[],
  error?: string,
  unconfirmed_food?: Food,
}

export type CalculateResponse = {
  items: ItemResult[],
  total: Nutrition,
}

export type LogItem = {
  id: number,
  food_id: number,
  food?: Food,
  quantity: number,
  unit: string,
}

export type MealLog = {
  id: number,
  user_id: number,
  logged_at: string,
  items: LogItem[],
}

export type DailySummary = {
  date: string,
  total: Nutrition,
  meals: MealLog[],
}

export type ParsedItem = {
  name: string,
  quantity: number,
  unit: string,
  preparation: string | null,
  brand: string | null,
}

export type ChatMealResponse = {
  parsed: ParsedItem[],
  result: CalculateResponse,
  needs_clarification: boolean,
}

export type NutritionEstimate = {
  name: string,
  calories: number,
  protein: number,
  carbs: number,
  fat: number,
  fiber: number,
  sodium: number,
  unit: string,
  unitquantity: number,
}
