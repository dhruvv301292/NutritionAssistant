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
						"brand":       map[string]any{"type": []string{"string", "null"}, "description": "e.g. 'Quest' — null if not specified"},
					},
					"required": []string{"name", "quantity", "unit", "preparation", "brand"},
				},
			},
		},
		Required: []string{"items"},
	}, parseMealToolName)

	message, err := c.anthropic.Messages.New(ctx, anthropic.MessageNewParams{
		Model:      anthropic.ModelClaudeSonnet4_5,
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

	message, err := c.anthropic.Messages.New(ctx, anthropic.MessageNewParams{
		Model:      anthropic.ModelClaudeSonnet4_5,
		MaxTokens:  1024,
		ToolChoice: anthropic.ToolChoiceParamOfTool(estimateNutritionToolName),
		Tools:      []anthropic.ToolUnionParam{tool},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"Estimate typical nutrition values for: " + foodName,
			)),
		},
	})
	if err != nil {
		return NutritionEstimate{}, fmt.Errorf("anthropic request failed: %w", err)
	}

	for _, block := range message.Content {
		if block.Type != "tool_use" || block.Name != estimateNutritionToolName {
			continue
		}
		var estimate NutritionEstimate
		if err := json.Unmarshal(block.Input, &estimate); err != nil {
			return NutritionEstimate{}, fmt.Errorf("failed to decode tool_use input: %w", err)
		}
		return estimate, nil
	}

	return NutritionEstimate{}, errNoToolUse
}
