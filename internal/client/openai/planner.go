package openai

import (
	"context"
	"fmt"
	"time"

	"github.com/dimdasci/seek/internal/models"
	"github.com/openai/openai-go"
	"go.uber.org/zap"
)

// PlanSearch builds a search plan for given query, returns it and an error if any.
func (c *Client) PlanSearch(ctx context.Context, query string) (*models.Plan, error) {

	// create a string with today's date
	today := fmt.Sprintf("%d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())
	prompt := fmt.Sprintf("%v\n\nToday is %v.\n\n<information_request>%v<information_request>", planningPrompt, today, query)

	// add timeout to the context
	ctx, cancel := context.WithTimeout(ctx, c.reasoningTimeout)
	defer cancel()

	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		Model:               openai.F(c.reasoningModel),
		MaxCompletionTokens: openai.Int(c.reasoningMaxTokens),
	})

	if err != nil {
		c.logger.Error("failed to get completion",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	// Log completion reason
	c.logger.Info("Search plan completion reason",
		zap.String("reason", string(chat.Choices[0].FinishReason)),
		zap.String("model", chat.Model),
		zap.Int64("completion tokens", chat.Usage.CompletionTokens),
		zap.Int64("max tokens", c.reasoningMaxTokens),
	)

	// create plan from chat response
	searchPlan, err := models.NewPlan(chat.Choices[0].Message.Content)
	if err != nil {
		c.logger.Error("failed to create plan from chat response",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	return searchPlan, nil
}
