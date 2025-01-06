package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dimdasci/seek/internal/models"
	"github.com/openai/openai-go"
	"go.uber.org/zap"
)

// CompileResults compiles the results of the search for the request.
// It returns a string with the compilation done with instructions.
func (c *Client) CompileResults(
	ctx context.Context,
	pages []models.Page,
	request *string,
	instructions *string,
) string {
	c.logger.Info("Compiling results", zap.String("request", *request))

	// gather key points from relevant pages
	keyPoints := c.getherKeyPoints(ctx, pages, request, instructions)

	// compile findings
	return c.compileFindings(keyPoints, *request, *instructions)
}

// getherKeyPoints gathers key points from relevant pages.
// It returns a string with the key points from all relevant pages.
func (c *Client) getherKeyPoints(
	ctx context.Context,
	pages []models.Page,
	request *string,
	instructions *string,
) string {
	var wg sync.WaitGroup
	results := make(chan string)
	done := make(chan struct{})

	// look for results in every page and keep only relevant
	for i, p := range pages {
		wg.Add(1)
		go func(i int, p models.Page) {
			defer wg.Done()
			relevant, keyPoints, err := c.analyzePage(ctx, &p, request, instructions)
			if err != nil {
				c.logger.Error("failed to analyze page",
					zap.Error(err),
					zap.String("url", p.URL),
					zap.String("title", p.Title),
					zap.Int("page_index", i))
				return
			}
			if relevant {
				results <- keyPoints
			}
		}(i, p)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	var compilation string

	for {
		select {
		case result := <-results:
			compilation += result + "\n\n"
		case <-done:
			close(results)
			return compilation
		}
	}
}

// analyzePage analyzes if the page contains information relevant to the request.
// It returns relevance and the key points from the page.
func (c *Client) analyzePage(
	ctx context.Context,
	page *models.Page,
	request *string,
	instructions *string,
) (relevant bool, keyPoints string, err error) {
	// create a string with today's date
	today := fmt.Sprintf("%d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())
	prompt := fmt.Sprintf("%v\n\n"+
		"Today is %v.\n\n"+
		"<title>%v<title>"+
		"<url>%v<url>"+
		"<content>%v<content>"+
		"<information_request>%v<information_request>"+
		"<compilation_instruction>%v<compilation_instruction>",
		relevanceUserPrompt,
		today,
		page.Title,
		page.URL,
		page.Content,
		request,
		instructions)

	c.logger.Debug("Relevance", zap.String("user_prompt", prompt))

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
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(c.analysisResultSchemaParam),
			},
		),
	})

	if err != nil {
		c.logger.Error("failed to analyze page",
			zap.Error(err),
			zap.String("url", page.URL),
			zap.String("title", page.Title))
		return false, "", err

	}

	// Log completion stats
	c.logger.Info("Page Analysis",
		zap.String("reason", string(chat.Choices[0].FinishReason)),
		zap.String("model", chat.Model),
		zap.Int64("input tokens", chat.Usage.PromptTokens),
		zap.Int64("completion tokens", chat.Usage.CompletionTokens),
		zap.Int64("max tokens", c.completionMaxTokens),
	)

	// create result from chat response
	result := AnalysisResult{}
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &result)
	if err != nil {
		c.logger.Error("failed to unmarshal chat response",
			zap.Error(err),
			zap.String("completion", chat.Choices[0].Message.Content))
		return false, "", err
	}

	c.logger.Debug("Page Analysis Done",
		zap.String("url", page.URL),
		zap.String("title", page.Title),
		zap.Bool("relevance", result.Relevance),
		zap.String("answer", result.Answer))

	return result.Relevance, result.Answer, nil
}

func (c *Client) compileFindings(results string, topic string, policy string) string {
	c.logger.Info("Compiling findings", zap.String("topic", topic))

	prompt := fmt.Sprintf("%v\n\n"+
		"<information_topic>%v<information_topic>"+
		"<search_results>%v<search_results>"+
		"<compilation_policy>%v<compilation_policy>",
		compilationPrompt,
		topic,
		results,
		policy)

	c.logger.Debug("Compilation", zap.String("prompt", prompt))

	// add timeout to the context
	ctx, cancel := context.WithTimeout(context.Background(), c.completionTimeout)
	defer cancel()

	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(relevanceSystemPrompt),
			openai.UserMessage(prompt),
		}),
		Model:               openai.F(c.completionModel),
		MaxCompletionTokens: openai.Int(c.completionMaxTokens),
		Temperature:         openai.Float(0.1),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(c.compilationResultSchemaParam),
			},
		),
	})

	if err != nil {
		c.logger.Error("failed to compile findings",
			zap.Error(err),
			zap.String("topic", topic))
		return ""
	}

	// Log completion stats
	c.logger.Info("Compilation",
		zap.String("reason", string(chat.Choices[0].FinishReason)),
		zap.String("model", chat.Model),
		zap.Int64("input tokens", chat.Usage.PromptTokens),
		zap.Int64("completion tokens", chat.Usage.CompletionTokens),
		zap.Int64("max tokens", c.completionMaxTokens))

	// create result from chat response
	result := CompilationResult{}
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &result)
	if err != nil {
		c.logger.Error("failed to unmarshal chat response",
			zap.Error(err),
			zap.String("completion", chat.Choices[0].Message.Content))
		return ""
	}

	c.logger.Debug("Compilation Done",
		zap.String("topic", topic),
		zap.String("compilation", result.Compilation))

	return result.Compilation
}
