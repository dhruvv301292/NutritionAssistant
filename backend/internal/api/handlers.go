package api

import (
	"encoding/json"
	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/dhruvv301292/nutrichat/internal/nutrition"
	"net/http"
	"fmt"
)

type Handler struct {
	foodRepo *foods.Repository
}

func NewHandler(foodRepo *foods.Repository) *Handler {
	return &Handler{foodRepo: foodRepo}
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "ok",
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) ListFoods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response, err := h.foodRepo.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list foods", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) SearchFoods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	queryFood := r.URL.Query().Get("q")
	response, err := h.foodRepo.Search(r.Context(), queryFood)
	if err != nil {
		http.Error(w, "failed to search foods", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CalculateNutrition(w http.ResponseWriter, r *http.Request) {
	var calcNutritionReq nutrition.CalculateNutritionRequest
	if err := json.NewDecoder(r.Body).Decode(&calcNutritionReq); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	foundFoods, err := h.foodRepo.Search(r.Context(), calcNutritionReq.FoodName)
	if err != nil {
		http.Error(w, "failed to search foods", http.StatusInternalServerError)
		return
	}
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