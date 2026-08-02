package main

import (
	"context"
	"log"
	"os"
	"github.com/dhruvv301292/nutrichat/internal/api"
	"github.com/dhruvv301292/nutrichat/internal/db"
	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/go-chi/chi/v5"
	"net/http"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}
	pool, err := db.NewPool(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	foodRepo := foods.NewRepository(pool)
	foodService := foods.NewService(foodRepo) 
	handler := api.NewHandler(foodService)

	r := chi.NewRouter()
	r.Get("/health", api.Health)
	r.Get("/foods", handler.ListFoods)
	r.Get("/foods/search", handler.SearchFoods)
	r.Post("/nutrition/calculate", handler.CalculateNutrition)
	http.ListenAndServe(":8080", r)
}
