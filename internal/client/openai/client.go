package openai

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.uber.org/zap"
)

type Client struct {
	client *openai.Client
	logger *zap.Logger
}

func NewClient(apiKey string, logger *zap.Logger) *Client {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Client{
		client: client,
		logger: logger,
	}
}

func (c *Client) SearchAnswer(ctx context.Context, query string) (string, error) {
	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(query),
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
	})

	if err != nil {
		c.logger.Error("failed to get completion",
			zap.Error(err),
			zap.String("query", query))
		return "", err
	}

	return chat.Choices[0].Message.Content, nil
}
