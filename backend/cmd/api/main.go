package main

import (
	"context"
	"log"
	"os"
	"github.com/dhruvv301292/nutrichat/internal/api"
	"github.com/dhruvv301292/nutrichat/internal/db"
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
	r := chi.NewRouter()
	r.Get("/health", api.Health)
	r.Get("/foods", api.ListFoods)
	r.Get("/foods/search", api.SearchFoods)
	r.Post("/nutrition/calculate", api.CalculateNutrition)
	http.ListenAndServe(":8080", r)
}
