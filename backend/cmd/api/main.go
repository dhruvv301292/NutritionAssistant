package main

import (
	"context"
	"log"
	"os"
	"strings"
	"github.com/dhruvv301292/nutrichat/internal/ai"
	"github.com/dhruvv301292/nutrichat/internal/api"
	"github.com/dhruvv301292/nutrichat/internal/auth"
	"github.com/dhruvv301292/nutrichat/internal/db"
	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/dhruvv301292/nutrichat/internal/fooddata"
	"github.com/dhruvv301292/nutrichat/internal/goals"
	"github.com/dhruvv301292/nutrichat/internal/meals"
	"github.com/dhruvv301292/nutrichat/internal/users"
	"github.com/go-chi/chi/v5"
	"net/http"
	"github.com/joho/godotenv"
)

// usdaProvider and fatSecretProvider adapt the concrete fooddata clients to
// foods.Service's externalLookup interface, translating *fooddata.Result to
// *foods.ExternalResult. Kept here rather than in fooddata so that package
// doesn't need to depend on foods.
type usdaProvider struct{ client *fooddata.USDAClient }

func (p usdaProvider) Lookup(ctx context.Context, query string) (*foods.ExternalResult, error) {
	return adaptResult(p.client.Lookup(ctx, query))
}

type fatSecretProvider struct{ client *fooddata.FatSecretClient }

func (p fatSecretProvider) Lookup(ctx context.Context, query string) (*foods.ExternalResult, error) {
	return adaptResult(p.client.Lookup(ctx, query))
}

func adaptResult(r *fooddata.Result, err error) (*foods.ExternalResult, error) {
	if err != nil || r == nil {
		return nil, err
	}
	return &foods.ExternalResult{
		Name:            r.Name,
		Calories:        r.Calories,
		Protein:         r.Protein,
		Carbs:           r.Carbs,
		Fat:             r.Fat,
		Fiber:           r.Fiber,
		Sodium:          r.Sodium,
		GramsPerServing: r.GramsPerServing,
	}, nil
}

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

	var externalProviders []foods.ExternalLookup
	if usdaKey := os.Getenv("USDA_FDC_API_KEY"); usdaKey != "" {
		externalProviders = append(externalProviders, usdaProvider{client: fooddata.NewUSDAClient(usdaKey)})
	}
	if fsID, fsSecret := os.Getenv("FATSECRET_CLIENT_ID"), os.Getenv("FATSECRET_CLIENT_SECRET"); fsID != "" && fsSecret != "" {
		externalProviders = append(externalProviders, fatSecretProvider{client: fooddata.NewFatSecretClient(fsID, fsSecret)})
	}
	foodService := foods.NewService(foodRepo, externalProviders...)

	mealRepo := meals.NewRepository(pool)
	matcher := meals.NewMatcher(foodService)
	mealService := meals.NewService(mealRepo, matcher, foodService)

	aiClient := ai.NewClient()

	goalsRepo := goals.NewRepository(pool)

	usersRepo := users.NewRepository(pool)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	authSigner := auth.NewSigner(jwtSecret)

	var googleAuds []string
	if auds := os.Getenv("GOOGLE_OAUTH_CLIENT_IDS"); auds != "" {
		googleAuds = strings.Split(auds, ",")
	}

	handler := api.NewHandler(foodService, mealService, aiClient, goalsRepo, usersRepo, authSigner, googleAuds)

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Get("/health", api.Health)
	r.Post("/auth/google", handler.GoogleLogin)
	r.Get("/foods", handler.ListFoods)
	r.Get("/foods/search", handler.SearchFoods)
	r.Post("/foods/estimate", handler.EstimateFood)
	r.Post("/foods", handler.CreateFood)
	r.Post("/nutrition/calculate", handler.CalculateNutrition)
	r.Post("/meals/calculate", handler.CalculateMeal)
	r.Post("/ai/parse-meal", handler.ParseMeal)
	r.Post("/chat/meal", handler.ChatMeal)

	r.Group(func(r chi.Router) {
		r.Use(authSigner.RequireAuth)
		r.Get("/me", handler.Me)
		r.Post("/meals", handler.SaveMeal)
		r.Get("/meals/today", handler.MealsToday)
		r.Get("/summary/daily", handler.DailySummary)
		r.Get("/goals", handler.GetGoals)
		r.Put("/goals", handler.PutGoals)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// corsMiddleware allows the mobile app and web frontend, running on
// different origins than the deployed API, to call it directly.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}
