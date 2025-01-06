/* Package openai provides a client for the OpenAI API. */
package openai

import (
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.uber.org/zap"
)

// Client is a client for the OpenAI API.
type Client struct {
	client              *openai.Client   // OpenAI API client
	logger              *zap.Logger      // Logger
	reasoningModel      openai.ChatModel // Model to use for reasoning
	completionModel     openai.ChatModel // Model to use for completion
	reasoningTimeout    time.Duration    // Timeout for reasoning
	completionTimeout   time.Duration    // Timeout for completion
	reasoningMaxTokens  int64            // Max tokens for reasoning
	completionMaxTokens int64            // Max tokens for completion

	analysisResultSchemaParam    openai.ResponseFormatJSONSchemaJSONSchemaParam // Schema for analysis result
	compilationResultSchemaParam openai.ResponseFormatJSONSchemaJSONSchemaParam // Schema for compilation result
}

// NewClient creates a new OpenAI API client with apiKey, and logger.
// It returns a pointer to the client.
func NewClient(
	apiKey string,
	logger *zap.Logger,
	reasoningModel openai.ChatModel,
	completionModel openai.ChatModel,
	reasoningTimeout time.Duration,
	completionTimeout time.Duration,
	reasoningMaxTokens int64,
	completionMaxTokens int64,
) (*Client, error) {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	rm, err := model(reasoningModel)
	if err != nil {
		logger.Fatal("failed to create reasoning model",
			zap.Error(err),
			zap.String("requested model", reasoningModel))
		return nil, err
	}

	sm, err := model(completionModel)
	if err != nil {
		logger.Fatal("failed to create service model",
			zap.Error(err),
			zap.String("requested model", completionModel))
		return nil, err
	}

	analysisResultSchemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("PageAnalysis"), // Updated to match the required pattern
		Description: openai.F("Relevance and key points from the page"),
		Schema:      openai.F(GenerateSchema[AnalysisResult]()),
		Strict:      openai.Bool(true),
	}
	compilationResultSchemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("PageAnalysis"), // Updated to match the required pattern
		Description: openai.F("Relevance and key points from the page"),
		Schema:      openai.F(GenerateSchema[CompilationResult]()),
		Strict:      openai.Bool(true),
	}

	return &Client{
		client:                       client,
		logger:                       logger,
		reasoningModel:               rm,
		completionModel:              sm,
		reasoningTimeout:             reasoningTimeout,
		completionTimeout:            completionTimeout,
		reasoningMaxTokens:           reasoningMaxTokens,
		completionMaxTokens:          completionMaxTokens,
		analysisResultSchemaParam:    analysisResultSchemaParam,
		compilationResultSchemaParam: compilationResultSchemaParam,
	}, nil
}

func model(m string) (openai.ChatModel, error) {
	switch m {
	case "gpt4", "gpt4o-mini":
		return openai.ChatModelGPT4oMini, nil
	case "o1-preview":
		return openai.ChatModelO1Preview, nil
	case "o1-mini":
		return openai.ChatModelO1Mini, nil
	case "gpt4o":
		return openai.ChatModelGPT4o, nil

	default:
		return "", fmt.Errorf("unknown model: %s", m)
	}
}
