package api

import (
	"encoding/json"
	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/dhruvv301292/nutrichat/internal/nutrition"
	"net/http"
	"fmt"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "ok",
	}
	json.NewEncoder(w).Encode(response)
}

func ListFoods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := foods.SeedFoods
	json.NewEncoder(w).Encode(response)
}

func SearchFoods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	queryFood := r.URL.Query().Get("q")
	response := foods.SearchFoods(queryFood)
	json.NewEncoder(w).Encode(response)
}

func CalculateNutrition(w http.ResponseWriter, r *http.Request) {
	var calcNutritionReq nutrition.CalculateNutritionRequest
	if err:= json.NewDecoder(r.Body).Decode(&calcNutritionReq); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
    	return
	}
	foundFoods := foods.SearchFoods(calcNutritionReq.FoodName)
	if len(foundFoods) == 0 {
		errorMessage := fmt.Sprintf("Food: %s not found", calcNutritionReq.FoodName)
		http.Error(w, errorMessage, http.StatusNotFound)
		return
	}
	calculatedNutrition, calculatedNutritionError := nutrition.CalculateNutrition(foundFoods[0], calcNutritionReq.Quantity, calcNutritionReq.Unit)
	if calculatedNutritionError != nil {
		errorMessage := fmt.Sprintf("Could not calculate nutrition for Food: %s", calcNutritionReq.FoodName)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(calculatedNutrition)
}