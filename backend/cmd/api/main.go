package main

import (
	"github.com/dhruvv301292/nutrichat/internal/api"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	r.Get("/health", api.Health)
	r.Get("/foods", api.ListFoods)
	r.Get("/foods/search", api.SearchFoods)
	r.Post("/nutrition/calculate", api.CalculateNutrition)
	http.ListenAndServe(":8080", r)
}
