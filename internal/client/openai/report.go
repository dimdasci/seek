package openai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"go.uber.org/zap"
)

// WriteReport writes the final report for the request.
// It returns a string with the final report.
func (c *Client) WriteReport(
	ctx context.Context,
	findings *string,
	request *string,
	plan *string,
	instructions *string,
) string {
	if findings == nil || request == nil || plan == nil || instructions == nil {
		c.logger.Error("One or more input parameters are nil")
		return "Failed to write report: one or more input parameters are nil"
	}

	c.logger.Info("Writing final report", zap.String("request", *request))

	prompt := fmt.Sprintf(
		"%s\n\n"+
			"<information_request>%s<information_request>\n\n"+
			"<plan>%s<plan>\n\n"+
			"<findings>%s<findings>\n\n"+
			"<instructions>%s<instructions>",
		finalReportPrompt,
		*request,
		*plan,
		*findings,
		*instructions,
	)

	// add timeout to the context
	ctx, cancel := context.WithTimeout(ctx, c.reasoningTimeout)
	defer cancel()

	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(relevanceSystemPrompt),
			openai.UserMessage(prompt),
		}),
		Model:               openai.F(c.completionModel),
		MaxCompletionTokens: openai.Int(c.completionMaxTokens),
		Temperature:         openai.Float(0.1),
	})

	if err != nil {
		c.logger.Error("Failed to write report", zap.Error(err))
		return fmt.Sprintf("Failed to write report: %v", err)
	}

	// compile findings
	return chat.Choices[0].Message.Content
}
