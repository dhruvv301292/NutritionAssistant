package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

const parseMealToolName = "record_meal_items"
const estimateNutritionToolName = "record_nutrition_estimate"
const verifyBrandMatchToolName = "record_brand_match_verdict"
const generateMealTitleToolName = "record_meal_title"

var errNoToolUse = errors.New("model did not return a tool_use block")

type Client struct {
	anthropic anthropic.Client
}

func NewClient() *Client {
	// anthropic.NewClient defaults to os.LookupEnv("ANTHROPIC_API_KEY")
	return &Client{anthropic: anthropic.NewClient()}
}

// ParseMeal turns a free-text meal description into structured items. The
// LLM's only responsibility is extraction — it never computes nutrition;
// that stays in the Go nutrition engine (see CLAUDE.md's Key Architecture
// Rule). Forcing tool_choice ensures the model must return arguments
// matching our schema rather than a free-form reply we'd have to parse
// ourselves.
func (c *Client) ParseMeal(ctx context.Context, text string) (ParsedMeal, error) {
	tool := anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":        map[string]any{"type": "string", "description": "the food's name, as plainly as possible — strip subjective/filler words (e.g. 'chicken thigh', not 'delicious chicken'), but KEEP words that distinguish one product variant from another, since e.g. 'diet'/'zero'/'zero sugar'/'light' change the nutrition values and must not be dropped. IMPORTANT: name and brand are concatenated as 'brand name' later — name must NEVER repeat the brand or any word/spelling that IS the brand under a different name (Coke/Coca-Cola, Coors/Coors, Bud/Budweiser are the same brand). Correct: brand='Coca-Cola', name='Zero Sugar' (for coke zero) or name='Diet' (for diet coke) or name='' (for plain coke, nothing to add beyond the brand itself). Wrong: name='Diet Coke' or name='Cola' when brand is already 'Coca-Cola' — that duplicates to 'Coca-Cola Diet Coke' / 'Coca-Cola Cola'. If brand is null, name stands alone and should include the generic category word (e.g. name='cola')."},
						"quantity":    map[string]any{"type": "number"},
						"unit":        map[string]any{"type": "string", "enum": []string{"grams", "ounces", "count"}, "description": "grams/ounces for weight, count for discrete items (eggs, bottles, bars) — always convert other units (cups, tbsp, etc.) to your best estimate in grams"},
						"preparation": map[string]any{"type": []string{"string", "null"}, "description": "e.g. 'raw', 'grilled', 'cooked' — null if not specified"},
						"brand":       map[string]any{"type": []string{"string", "null"}, "description": "e.g. 'Quest'. Expand common brand abbreviations/initialisms to the full brand name (e.g. 'ON' or 'on' before a protein product means 'Optimum Nutrition', not the word \"on\" — don't fold it into name or drop it as a filler word). Null only if no brand is mentioned at all."},
					},
					"required": []string{"name", "quantity", "unit", "preparation", "brand"},
				},
			},
		},
		Required: []string{"items"},
	}, parseMealToolName)
	// Cache breakpoint: this tool schema is identical on every call, so
	// caching it means only the user's meal text (a few dozen tokens) is
	// billed at full price — the schema itself is billed once per cache
	// window (default 5 min) instead of on every request.
	tool.OfTool.CacheControl = anthropic.NewCacheControlEphemeralParam()

	// Haiku, not Sonnet: this is constrained extraction against a forced
	// tool schema, not open-ended reasoning — cheaper model, same task.
	message, err := c.anthropic.Messages.New(ctx, anthropic.MessageNewParams{
		Model:      anthropic.ModelClaudeHaiku4_5,
		MaxTokens:  1024,
		ToolChoice: anthropic.ToolChoiceParamOfTool(parseMealToolName),
		Tools:      []anthropic.ToolUnionParam{tool},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(text)),
		},
	})
	if err != nil {
		return ParsedMeal{}, fmt.Errorf("anthropic request failed: %w", err)
	}

	for _, block := range message.Content {
		if block.Type != "tool_use" || block.Name != parseMealToolName {
			continue
		}
		var meal ParsedMeal
		if err := json.Unmarshal(block.Input, &meal); err != nil {
			return ParsedMeal{}, fmt.Errorf("failed to decode tool_use input: %w", err)
		}
		return meal, nil
	}

	return ParsedMeal{}, errNoToolUse
}

// EstimateNutrition asks the model to guess a food's typical nutrition
// values when it couldn't be found in our database or any external food
// API. This is the one place the LLM produces numbers that look like
// nutrition data — but it's a draft, not a calculation: callers must
// surface it as an editable suggestion and never persist or use it in a
// meal total until a human confirms or corrects it (see CLAUDE.md's Key
// Architecture Rule).
// brand, if non-empty, is folded into the prompt (so the model knows which
// specific branded product to search for) and stamped onto the returned
// estimate directly — not something the model fills into the tool schema
// itself, since it's already a known, structured value from the caller
// (the parsed meal item or an existing food match), not something to guess.
func (c *Client) EstimateNutrition(ctx context.Context, foodName, brand string) (NutritionEstimate, error) {
	tool := anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"name":         map[string]any{"type": "string", "description": "cleaned-up food name"},
			"calories":     map[string]any{"type": "number", "description": "typical calories for the given unit/unit_quantity"},
			"protein":      map[string]any{"type": "number", "description": "grams"},
			"carbs":        map[string]any{"type": "number", "description": "grams"},
			"fat":          map[string]any{"type": "number", "description": "grams"},
			"fiber":        map[string]any{"type": "number", "description": "grams"},
			"sodium":       map[string]any{"type": "number", "description": "milligrams"},
			"unit":         map[string]any{"type": "string", "enum": []string{"grams", "count"}, "description": "grams for a 100g basis, count for a discrete item/serving/package"},
			"unitquantity": map[string]any{"type": "number", "description": "must be exactly 100 when unit is \"grams\". When unit is \"count\", this is how many individual pieces the values below cover — use the food's REAL official serving/package size (e.g. 8 for an \"8 piece\" nugget order, 12 for a \"12 piece\", 30 for a \"30 piece\"), never force it to 1 for a multi-piece food. Only use 1 when the food is naturally a single discrete unit (one egg, one bar, one can). Never 0."},
		},
		Required: []string{"name", "calories", "protein", "carbs", "fat", "fiber", "sodium", "unit", "unitquantity"},
	}, estimateNutritionToolName)
	// Cache breakpoints on both the tool schema and the system instructions
	// below: neither changes between calls, only the food name does. Once
	// cached, a repeat estimate request only pays full price for the short
	// food-name user message.
	tool.OfTool.CacheControl = anthropic.NewCacheControlEphemeralParam()

	system := anthropic.TextBlockParam{
		Text: "Estimate typical nutrition values for the given food. If a " +
			"preparation state (raw/cooked/grilled/etc.) is mentioned, use " +
			"values for that specific state — cooked and raw nutrition profiles " +
			"differ meaningfully (cooking concentrates most nutrients per gram via " +
			"water loss). If the food is a branded product or a chain restaurant " +
			"item, use the web_search tool to check the brand's or chain's official " +
			"nutrition page for exact published values before estimating — don't " +
			"rely on memory alone for these, since exact values matter and change " +
			"over time. For unbranded generic foods, a web search usually isn't " +
			"necessary. Once you have the values (searched or estimated), call " +
			estimateNutritionToolName + " with the result.",
	}
	system.CacheControl = anthropic.NewCacheControlEphemeralParam()

	webSearch := anthropic.ToolUnionParam{OfWebSearchTool20250305: &anthropic.WebSearchTool20250305Param{
		MaxUses: anthropic.Int(3),
	}}

	userText := foodName
	if brand != "" {
		userText = brand + " " + foodName
	}

	// ToolChoice is left at the default (auto) rather than forced onto
	// estimateNutritionToolName, since forcing it would prevent Claude from
	// ever calling web_search first — it needs the freedom to search, then
	// call our tool with what it found. web_search itself runs server-side
	// within this single request/response; no manual tool-result loop
	// needed here.
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 2048,
		System:    []anthropic.TextBlockParam{system},
		Tools:     []anthropic.ToolUnionParam{tool, webSearch},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userText)),
		},
	}

	message, err := c.anthropic.Messages.New(ctx, params)
	if err != nil {
		return NutritionEstimate{}, fmt.Errorf("anthropic request failed: %w", err)
	}
	if estimate, ok := extractEstimate(message); ok {
		return withBrand(estimate, brand), nil
	}

	// auto tool_choice occasionally lets the model stop after searching
	// without ever calling our tool (e.g. it ran out of budget mid-search,
	// or decided the search results were the final answer). Retry once,
	// forcing the tool directly with no web_search available — worse
	// accuracy for that one retry, but still correct behavior (an
	// estimate) rather than a hard failure on what's now the primary path.
	params.ToolChoice = anthropic.ToolChoiceParamOfTool(estimateNutritionToolName)
	params.Tools = []anthropic.ToolUnionParam{tool}
	message, err = c.anthropic.Messages.New(ctx, params)
	if err != nil {
		return NutritionEstimate{}, fmt.Errorf("anthropic request failed: %w", err)
	}
	if estimate, ok := extractEstimate(message); ok {
		return withBrand(estimate, brand), nil
	}

	return NutritionEstimate{}, errNoToolUse
}

func withBrand(estimate NutritionEstimate, brand string) NutritionEstimate {
	if brand != "" {
		estimate.Brand = &brand
	}
	return estimate
}

// GenerateMealTitle comes up with a short, fun name for a logged meal based
// on its ingredient list (e.g. "diet coke, whey protein powder" -> "Whey
// Coke Float"). Purely cosmetic — never affects matching, calculation, or
// anything else in the nutrition pipeline — so a forced-schema Haiku call
// is enough, same tier as VerifyBrandMatch.
func (c *Client) GenerateMealTitle(ctx context.Context, ingredients []string) (string, error) {
	tool := anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"title": map[string]any{"type": "string", "description": "a short, fun, punny title (2-5 words) for a meal made of these ingredients — playful, not literal, no quotes around it"},
		},
		Required: []string{"title"},
	}, generateMealTitleToolName)

	prompt := "Ingredients: " + strings.Join(ingredients, ", ") +
		"\n\nCome up with a short, fun title for this meal."

	message, err := c.anthropic.Messages.New(ctx, anthropic.MessageNewParams{
		Model:      anthropic.ModelClaudeHaiku4_5,
		MaxTokens:  256,
		ToolChoice: anthropic.ToolChoiceParamOfTool(generateMealTitleToolName),
		Tools:      []anthropic.ToolUnionParam{tool},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}

	for _, block := range message.Content {
		if block.Type != "tool_use" || block.Name != generateMealTitleToolName {
			continue
		}
		var result struct {
			Title string `json:"title"`
		}
		if err := json.Unmarshal(block.Input, &result); err != nil {
			return "", fmt.Errorf("failed to decode tool_use input: %w", err)
		}
		return result.Title, nil
	}

	return "", errNoToolUse
}

// VerifyBrandMatch asks the model whether a candidate database row is the
// same branded product the user asked for. This is a last-resort tiebreaker
// for the case where trigram brand similarity falls in an ambiguous middle
// zone (see foods/repository.go's brandAmbiguousLow/High) — not confidently
// a match, not confidently a mismatch (e.g. "opti nutrition" vs "optimum
// nutrition", or a minor misspelling). The model only answers yes/no on a
// specific pairing; it never generates nutrition values here, so it can't
// introduce the kind of confidently-wrong data EstimateNutrition is
// explicitly allowed to (and which still goes through human review).
func (c *Client) VerifyBrandMatch(ctx context.Context, queryBrand, queryProduct, candidateBrand, candidateName string) (bool, error) {
	tool := anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"same_product": map[string]any{"type": "boolean", "description": "true only if the candidate is the same specific branded product the user asked for — not just the same category or a similar product from a different brand"},
		},
		Required: []string{"same_product"},
	}, verifyBrandMatchToolName)

	prompt := fmt.Sprintf(
		"User asked for: %q (brand: %q)\nDatabase candidate: %q (brand: %q)\n\n"+
			"Is the database candidate the same specific branded product the user asked "+
			"for? Answer false if it's a different brand, even if the product type "+
			"matches (e.g. a different brand's protein shake is NOT a match).",
		queryProduct, queryBrand, candidateName, candidateBrand,
	)

	// Haiku, not Sonnet: a forced-schema yes/no verdict on a specific
	// pairing, not open-ended reasoning.
	message, err := c.anthropic.Messages.New(ctx, anthropic.MessageNewParams{
		Model:      anthropic.ModelClaudeHaiku4_5,
		MaxTokens:  256,
		ToolChoice: anthropic.ToolChoiceParamOfTool(verifyBrandMatchToolName),
		Tools:      []anthropic.ToolUnionParam{tool},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return false, fmt.Errorf("anthropic request failed: %w", err)
	}

	for _, block := range message.Content {
		if block.Type != "tool_use" || block.Name != verifyBrandMatchToolName {
			continue
		}
		var verdict struct {
			SameProduct bool `json:"same_product"`
		}
		if err := json.Unmarshal(block.Input, &verdict); err != nil {
			return false, fmt.Errorf("failed to decode tool_use input: %w", err)
		}
		return verdict.SameProduct, nil
	}

	return false, errNoToolUse
}

func extractEstimate(message *anthropic.Message) (NutritionEstimate, bool) {
	for _, block := range message.Content {
		if block.Type != "tool_use" || block.Name != estimateNutritionToolName {
			continue
		}
		var estimate NutritionEstimate
		if err := json.Unmarshal(block.Input, &estimate); err != nil {
			return NutritionEstimate{}, false
		}
		return estimate, true
	}
	return NutritionEstimate{}, false
}
