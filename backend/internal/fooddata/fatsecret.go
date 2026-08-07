package fooddata

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	fatSecretTokenURL  = "https://oauth.fatsecret.com/connect/token"
	fatSecretSearchURL = "https://platform.fatsecret.com/rest/foods/search/v1"
	fatSecretDetailURL = "https://platform.fatsecret.com/rest/food/v4"
)

type FatSecretClient struct {
	clientID     string
	clientSecret string
	http         *http.Client

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

func NewFatSecretClient(clientID, clientSecret string) *FatSecretClient {
	return &FatSecretClient{clientID: clientID, clientSecret: clientSecret, http: &http.Client{}}
}

// token returns a cached access token, fetching a new one via the
// client-credentials flow if none is cached or the cached one has expired.
func (c *FatSecretClient) token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return c.accessToken, nil
	}

	body := strings.NewReader("grant_type=client_credentials&scope=basic")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fatSecretTokenURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("fatsecret token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fatsecret token request failed: status %d", resp.StatusCode)
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("fatsecret token decode failed: %w", err)
	}

	c.accessToken = parsed.AccessToken
	// Refresh a minute early so a near-expiry token isn't used for a
	// request that then fails partway through.
	c.expiresAt = time.Now().Add(time.Duration(parsed.ExpiresIn)*time.Second - time.Minute)
	return c.accessToken, nil
}

// oneOrMany decodes a field that FatSecret sometimes returns as a bare
// object (single result) and sometimes as an array (multiple results) —
// the same shape ambiguity applies to both the food search list and a
// food's serving list.
type oneOrMany[T any] []T

func (o *oneOrMany[T]) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var many []T
		if err := json.Unmarshal(data, &many); err != nil {
			return err
		}
		*o = many
		return nil
	}
	var single T
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	*o = oneOrMany[T]{single}
	return nil
}

type fatSecretSearchResponse struct {
	Foods struct {
		Food oneOrMany[fatSecretFoodSummary] `json:"food"`
	} `json:"foods"`
}

type fatSecretFoodSummary struct {
	FoodID   string `json:"food_id"`
	FoodName string `json:"food_name"`
}

type fatSecretDetailResponse struct {
	Food struct {
		FoodName string `json:"food_name"`
		Servings struct {
			Serving oneOrMany[fatSecretServing] `json:"serving"`
		} `json:"servings"`
	} `json:"food"`
}

type fatSecretServing struct {
	MetricServingAmount string `json:"metric_serving_amount"`
	MetricServingUnit   string `json:"metric_serving_unit"`
	Calories            string `json:"calories"`
	Protein             string `json:"protein"`
	Carbohydrate        string `json:"carbohydrate"`
	Fat                 string `json:"fat"`
	Fiber               string `json:"fiber"`
	Sodium              string `json:"sodium"`
}

// Lookup searches FatSecret for query (branded/restaurant foods are its
// strength) and returns the first match's structured nutrition, normalized
// to Result. FatSecret returns numeric fields as strings and its serving
// size varies per food, so GramsPerServing is set whenever the serving's
// unit is grams.
func (c *FatSecretClient) Lookup(ctx context.Context, query string) (*Result, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}

	searchParams := url.Values{}
	searchParams.Set("search_expression", query)
	searchParams.Set("format", "json")
	searchParams.Set("max_results", "1")

	var searchResult fatSecretSearchResponse
	if err := c.getJSON(ctx, fatSecretSearchURL+"?"+searchParams.Encode(), token, &searchResult); err != nil {
		return nil, err
	}
	if len(searchResult.Foods.Food) == 0 {
		return nil, nil
	}
	foodID := searchResult.Foods.Food[0].FoodID

	detailParams := url.Values{}
	detailParams.Set("food_id", foodID)
	detailParams.Set("format", "json")

	var detail fatSecretDetailResponse
	if err := c.getJSON(ctx, fatSecretDetailURL+"?"+detailParams.Encode(), token, &detail); err != nil {
		return nil, err
	}
	if len(detail.Food.Servings.Serving) == 0 {
		return nil, nil
	}
	serving := detail.Food.Servings.Serving[0]

	result := &Result{
		Name:     detail.Food.FoodName,
		Calories: parseFloatOrZero(serving.Calories),
		Protein:  parseFloatOrZero(serving.Protein),
		Carbs:    parseFloatOrZero(serving.Carbohydrate),
		Fat:      parseFloatOrZero(serving.Fat),
		Fiber:    parseFloatOrZero(serving.Fiber),
		Sodium:   parseFloatOrZero(serving.Sodium),
	}
	if serving.MetricServingUnit == "g" {
		grams := parseFloatOrZero(serving.MetricServingAmount)
		if grams > 0 {
			result.GramsPerServing = &grams
		}
	}

	return result, nil
}

func (c *FatSecretClient) getJSON(ctx context.Context, requestURL, token string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("fatsecret request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fatsecret request failed: status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("fatsecret response decode failed: %w", err)
	}
	return nil
}

func parseFloatOrZero(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
