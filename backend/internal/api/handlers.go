package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dhruvv301292/nutrichat/internal/ai"
	"github.com/dhruvv301292/nutrichat/internal/auth"
	"github.com/dhruvv301292/nutrichat/internal/foods"
	"github.com/dhruvv301292/nutrichat/internal/goals"
	"github.com/dhruvv301292/nutrichat/internal/meals"
	"github.com/dhruvv301292/nutrichat/internal/nutrition"
	"github.com/dhruvv301292/nutrichat/internal/users"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	foodService *foods.Service
	mealService *meals.Service
	aiClient    *ai.Client
	goalsRepo   *goals.Repository
	usersRepo   *users.Repository
	authSigner  *auth.Signer
	googleAuds  []string
	appleAuds   []string
}

func NewHandler(food *foods.Service, meal *meals.Service, aiClient *ai.Client, goalsRepo *goals.Repository, usersRepo *users.Repository, authSigner *auth.Signer, googleAuds []string, appleAuds []string) *Handler {
	return &Handler{
		foodService: food,
		mealService: meal,
		aiClient:    aiClient,
		goalsRepo:   goalsRepo,
		usersRepo:   usersRepo,
		authSigner:  authSigner,
		googleAuds:  googleAuds,
		appleAuds:   appleAuds,
	}
}

type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
}

type GoogleLoginResponse struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

// GoogleLogin verifies a Google-issued ID token (obtained client-side via
// expo-auth-session), upserts the corresponding user, and issues our own
// session JWT — the token the client sends on every subsequent request.
// Google is only consulted here, not on the hot path.
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		http.Error(w, "id_token is required", http.StatusBadRequest)
		return
	}

	claims, err := auth.VerifyGoogleIDToken(r.Context(), req.IDToken, h.googleAuds)
	if err != nil {
		log.Printf("GoogleLogin: %v", err)
		http.Error(w, "invalid google id token", http.StatusUnauthorized)
		return
	}

	user, err := h.usersRepo.UpsertGoogleUser(r.Context(), claims.Sub, claims.Email, claims.Name)
	if err != nil {
		log.Printf("GoogleLogin: upsert user: %v", err)
		http.Error(w, "failed to sign in", http.StatusInternalServerError)
		return
	}

	token, err := h.authSigner.IssueSession(user.ID)
	if err != nil {
		log.Printf("GoogleLogin: issue session: %v", err)
		http.Error(w, "failed to sign in", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(GoogleLoginResponse{Token: token, User: user})
}

type AppleLoginRequest struct {
	IDToken string `json:"id_token"`
	// Name is only present client-side on the user's very first-ever Apple
	// sign-in — Apple includes it in the initial authorization credential,
	// not in the identity token itself, and never resends it afterward.
	Name string `json:"name"`
}

type AppleLoginResponse struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

// AppleLogin verifies an Apple-issued identity token (obtained client-side
// via expo-apple-authentication), upserts the corresponding user, and
// issues our own session JWT — same shape as GoogleLogin.
func (h *Handler) AppleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req AppleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		http.Error(w, "id_token is required", http.StatusBadRequest)
		return
	}

	claims, err := auth.VerifyAppleIDToken(r.Context(), req.IDToken, h.appleAuds)
	if err != nil {
		log.Printf("AppleLogin: %v", err)
		http.Error(w, "invalid apple id token", http.StatusUnauthorized)
		return
	}

	user, err := h.usersRepo.UpsertAppleUser(r.Context(), claims.Sub, claims.Email, req.Name)
	if err != nil {
		log.Printf("AppleLogin: upsert user: %v", err)
		http.Error(w, "failed to sign in", http.StatusInternalServerError)
		return
	}

	token, err := h.authSigner.IssueSession(user.ID)
	if err != nil {
		log.Printf("AppleLogin: issue session: %v", err)
		http.Error(w, "failed to sign in", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(AppleLoginResponse{Token: token, User: user})
}

// Me returns the authenticated user's profile — used on app cold start to
// resolve a stored session token back into a User without re-running the
// Google sign-in flow.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	user, err := h.usersRepo.FindByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
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
	response, err := h.foodService.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list foods", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) SearchFoods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	queryFood := r.URL.Query().Get("q")
	response, err := h.foodService.Search(r.Context(), queryFood)
	if err != nil {
		log.Printf("SearchFoods: %v", err)
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
	foundFoods, err := h.foodService.Search(r.Context(), calcNutritionReq.FoodName)
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

func (h *Handler) CalculateMeal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req meals.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "items must not be empty", http.StatusBadRequest)
		return
	}
	response := h.mealService.Calculate(r.Context(), req.Items)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) SaveMeal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	var req meals.SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "items must not be empty", http.StatusBadRequest)
		return
	}
	log, itemResults, err := h.mealService.Save(r.Context(), userID, req.Slot, req.Items)
	if err != nil {
		if errors.Is(err, meals.ErrUnresolvedItems) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "one or more items could not be resolved",
				"items": itemResults,
			})
			return
		}
		if errors.Is(err, meals.ErrInvalidSlot) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to save meal", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(log)
}

func (h *Handler) MealsToday(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	logs, err := h.mealService.Today(r.Context(), userID, tzOffsetMinutes(r))
	if err != nil {
		http.Error(w, "failed to fetch meals", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(logs)
}

// tzOffsetMinutes reads the client's local-vs-UTC offset (JS
// Date.getTimezoneOffset() convention: minutes local is BEHIND UTC), so
// "today"/daily-summary queries use the user's local calendar day instead
// of the server's UTC day. Missing/invalid values default to UTC (0).
func tzOffsetMinutes(r *http.Request) int {
	raw := r.URL.Query().Get("tz_offset_minutes")
	if raw == "" {
		return 0
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return v
}

func (h *Handler) DailySummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date query parameter is required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	summary, err := h.mealService.DailySummary(r.Context(), userID, date, tzOffsetMinutes(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(summary)
}

type ParseMealRequest struct {
	Text string `json:"text"`
}

func (h *Handler) ParseMeal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req ParseMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Text == "" {
		http.Error(w, "text must not be empty", http.StatusBadRequest)
		return
	}
	parsed, err := h.aiClient.ParseMeal(r.Context(), req.Text)
	if err != nil {
		http.Error(w, "failed to parse meal text", http.StatusBadGateway)
		return
	}
	json.NewEncoder(w).Encode(parsed)
}

type EstimateFoodRequest struct {
	Name  string `json:"name"`
	Brand string `json:"brand,omitempty"`
}

// EstimateFood is the fallback when a food isn't in our own database (see
// ai.Client.EstimateNutrition): the LLM's guess is returned as a draft
// only — nothing is persisted here.
// The client shows it as an editable form; POST /foods with the
// user-confirmed values is what actually saves it.
func (h *Handler) EstimateFood(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req EstimateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name must not be empty", http.StatusBadRequest)
		return
	}
	estimate, err := h.aiClient.EstimateNutrition(r.Context(), req.Name, req.Brand)
	if err != nil {
		http.Error(w, "failed to estimate nutrition", http.StatusBadGateway)
		return
	}
	json.NewEncoder(w).Encode(estimate)
}

// CreateFood persists a food the user confirmed (typically after editing an
// EstimateFood draft), so it's available for future meal logging like any
// other food in the database.
func (h *Handler) CreateFood(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var f nutrition.Food
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if f.Name == "" || f.Unit == "" || f.UnitQuantity <= 0 {
		http.Error(w, "name, unit, and a positive unit_quantity are required", http.StatusBadRequest)
		return
	}
	created, err := h.foodService.Create(r.Context(), f)
	if err != nil {
		http.Error(w, "failed to save food", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(created)
}

type ChatMealRequest struct {
	Text string `json:"text"`
}

type ChatMealResponse struct {
	Parsed             []ai.ParsedItem         `json:"parsed"`
	Result             meals.CalculateResponse `json:"result"`
	NeedsClarification bool                    `json:"needs_clarification"`
}

// needsClarification is true if any item couldn't be resolved cleanly —
// ambiguous matches or lookup failures — so the client knows to prompt the
// user with a follow-up rather than presenting partial/missing totals as
// if they were complete.
func needsClarification(items []meals.ItemResult) bool {
	for _, item := range items {
		if item.Ambiguous || item.Error != "" || item.UnconfirmedFood != nil {
			return true
		}
	}
	return false
}

// ChatMeal is the AI path end to end: free text in, structured nutrition
// out. The LLM only extracts items (ai.Client.ParseMeal); matching foods,
// converting units, and computing nutrition all happen in mealService,
// same deterministic engine the plain /meals/calculate endpoint uses.
func (h *Handler) ChatMeal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req ChatMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Text == "" {
		http.Error(w, "text must not be empty", http.StatusBadRequest)
		return
	}

	parsed, err := h.aiClient.ParseMeal(r.Context(), req.Text)
	if err != nil {
		http.Error(w, "failed to parse meal text", http.StatusBadGateway)
		return
	}
	if len(parsed.Items) == 0 {
		http.Error(w, "no food items found in text", http.StatusUnprocessableEntity)
		return
	}

	items := make([]meals.ItemRequest, len(parsed.Items))
	for i, p := range parsed.Items {
		items[i] = meals.ItemRequest{FoodName: p.Name, Quantity: p.Quantity, Unit: p.Unit, Preparation: p.Preparation, Brand: p.Brand}
	}

	result := h.mealService.Calculate(r.Context(), items)
	json.NewEncoder(w).Encode(ChatMealResponse{
		Parsed:             parsed.Items,
		Result:             result,
		NeedsClarification: needsClarification(result.Items),
	})
}

func (h *Handler) GetGoals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	g, err := h.goalsRepo.Get(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to fetch goals", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(g)
}

func (h *Handler) PutGoals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	var g goals.Goals
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	for _, m := range g.TrackedMacros {
		if !goals.ValidTrackedMacros[m] {
			http.Error(w, "invalid tracked_macros value: "+m, http.StatusBadRequest)
			return
		}
	}
	g.UserID = userID
	saved, err := h.goalsRepo.Upsert(r.Context(), g)
	if err != nil {
		http.Error(w, "failed to save goals", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(saved)
}
