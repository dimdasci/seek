/* Package openai provides a client for the OpenAI API. */
package openai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
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

// SearchAnswer searches for an answer to the query using the OpenAI API.
// It returns the answer as a string.
func (c *Client) SearchAnswer(ctx context.Context, query string) (string, error) {

	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf("%v\n\n<information_request>%v<information_request>", planningPrompt, query)),
		}),
		Model: openai.F(reasoningModel),
	})

	if err != nil {
		c.logger.Error("failed to get completion",
			zap.Error(err),
			zap.String("query", query))
		return "", err
	}

	return chat.Choices[0].Message.Content, nil
}
