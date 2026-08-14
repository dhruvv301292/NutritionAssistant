package main

import (
	"context"
	"github.com/dhruvv301292/nutrichat/internal/ai"
	"github.com/dhruvv301292/nutrichat/internal/api"
	"github.com/dhruvv301292/nutrichat/internal/auth"
	"github.com/dhruvv301292/nutrichat/internal/db"
	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/dhruvv301292/nutrichat/internal/goals"
	"github.com/dhruvv301292/nutrichat/internal/meals"
	"github.com/dhruvv301292/nutrichat/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"
)

// aiEstimatorProvider adapts ai.Client.EstimateNutrition to foods.Service's
// ExternalLookup interface. USDA/FatSecret were dropped from this path
// entirely — both returned confidently wrong matches for anything outside
// generic staple foods (a "David protein bar" search matched an unrelated
// South Beach bar; "jasmine rice" matched rice crackers) with no reliable
// way to detect the mismatch before it reached the user. The LLM estimate
// was already the last-resort fallback and already goes through the same
// human-review-before-save gate (see foods.Service.Create), so promoting it
// to the only source keeps that safety property while dropping a source
// that was silently wrong more often than it was silently right.
type aiEstimatorProvider struct{ client *ai.Client }

func (p aiEstimatorProvider) Lookup(ctx context.Context, query string) (*foods.ExternalResult, error) {
	// brand is already folded into query by the time it reaches here (see
	// Matcher.resolveBrandedItem's brand+" "+productName), so nothing extra
	// to pass — this adapter's caller doesn't have a separate structured
	// brand value at this point.
	estimate, err := p.client.EstimateNutrition(ctx, query, "")
	if err != nil {
		return nil, err
	}
	return &foods.ExternalResult{
		Name:            estimate.Name,
		Calories:        estimate.Calories,
		Protein:         estimate.Protein,
		Carbs:           estimate.Carbs,
		Fat:             estimate.Fat,
		Fiber:           estimate.Fiber,
		Sodium:          estimate.Sodium,
		Unit:            estimate.Unit,
		UnitQuantity:    estimate.UnitQuantity,
		GramsPerServing: estimate.GramsPerUnit,
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

	aiClient := ai.NewClient()

	foodService := foods.NewService(foodRepo, aiEstimatorProvider{client: aiClient})

	mealRepo := meals.NewRepository(pool)
	matcher := meals.NewMatcher(foodService, aiClient)
	mealService := meals.NewService(mealRepo, matcher, foodService, aiClient)

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

	var appleAuds []string
	if auds := os.Getenv("APPLE_OAUTH_AUDIENCES"); auds != "" {
		appleAuds = strings.Split(auds, ",")
	}

	handler := api.NewHandler(foodService, mealService, aiClient, goalsRepo, usersRepo, authSigner, googleAuds, appleAuds)

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Get("/health", api.Health)
	r.Post("/auth/google", handler.GoogleLogin)
	r.Post("/auth/apple", handler.AppleLogin)
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
