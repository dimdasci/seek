/* Package openai provides a client for the OpenAI API. */
package openai

import (
	"context"
	"fmt"
	"time"

	"github.com/dimdasci/seek/internal/model/plan"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Client is a client for the OpenAI API.
type Client struct {
	client *openai.Client // OpenAI API client
	logger *zap.Logger    // Logger
}

const (
	reasoningModel = openai.ChatModelO1Mini // Model to use for reasoning
	// serviceModel   = openai.ChatModelGPT4oMini // Model to use for service
)

// NewClient creates a new OpenAI API client with apiKey, and logger.
// It returns a pointer to the client.
func NewClient(apiKey string, logger *zap.Logger) *Client {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Client{
		client: client,
		logger: logger,
	}
}

// PlanSearch builds a search plan for given query, returns it and an error if any.
func (c *Client) PlanSearch(ctx context.Context, query string) (*plan.Plan, error) {

	// create a string with today's date
	today := fmt.Sprintf("%d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())
	prompt := fmt.Sprintf("%v\n\nToday is %v.\n\n<information_request>%v<information_request>", planningPrompt, today, query)

	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		Model:               openai.F(reasoningModel),
		MaxCompletionTokens: openai.Int(viper.GetInt64("openai.reasoning.max_tokens")),
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
		zap.Int64("max tokens", viper.GetInt64("openai.reasoning.max_tokens")),
	)

	// create plan from chat response
	searchPlan, err := plan.New(chat.Choices[0].Message.Content)
	if err != nil {
		c.logger.Error("failed to create plan from chat response",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	return searchPlan, nil
}
