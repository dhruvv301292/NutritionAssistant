package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

const parseMealToolName = "record_meal_items"
const estimateNutritionToolName = "record_nutrition_estimate"
const verifyBrandMatchToolName = "record_brand_match_verdict"

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
						"name":        map[string]any{"type": "string", "description": "the food's name, as plainly as possible (e.g. 'chicken thigh', not 'delicious chicken')"},
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
func (c *Client) EstimateNutrition(ctx context.Context, foodName string) (NutritionEstimate, error) {
	tool := anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"name":          map[string]any{"type": "string", "description": "cleaned-up food name"},
			"calories":      map[string]any{"type": "number", "description": "typical calories for the given unit/unit_quantity"},
			"protein":       map[string]any{"type": "number", "description": "grams"},
			"carbs":         map[string]any{"type": "number", "description": "grams"},
			"fat":           map[string]any{"type": "number", "description": "grams"},
			"fiber":         map[string]any{"type": "number", "description": "grams"},
			"sodium":        map[string]any{"type": "number", "description": "milligrams"},
			"unit":         map[string]any{"type": "string", "enum": []string{"grams", "count"}, "description": "grams for a 100g basis, count for a single typical serving/item"},
			"unitquantity": map[string]any{"type": "number", "enum": []float64{100, 1}, "description": "must be exactly 100 when unit is \"grams\", or exactly 1 when unit is \"count\" — never 0"},
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
			anthropic.NewUserMessage(anthropic.NewTextBlock(foodName)),
		},
	}

	message, err := c.anthropic.Messages.New(ctx, params)
	if err != nil {
		return NutritionEstimate{}, fmt.Errorf("anthropic request failed: %w", err)
	}
	if estimate, ok := extractEstimate(message); ok {
		return estimate, nil
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
		return estimate, nil
	}

	return NutritionEstimate{}, errNoToolUse
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
