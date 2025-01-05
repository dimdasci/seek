package webread

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dimdasci/seek/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// TavilyReadService provides web reading functionality.
type TavilyReadService struct {
	APIKey  string
	BaseURL string
	logger  *zap.Logger
}

// NewTavilyReadService creates a new instance of TavilyReadService.
func NewTavilyReadService(logger *zap.Logger) *TavilyReadService {
	return &TavilyReadService{
		APIKey:  viper.GetString("websearch.tavily.api_key"),
		BaseURL: viper.GetString("websearch.tavily.extract_url"),
		logger:  logger,
	}
}

// Read sends a request to Tavily's extraction API for the given URLs.
// It returns the web pages or an error.
func (t *TavilyReadService) Read(ctx context.Context, urls []string) (*models.WebPages, error) {
	requestBody, err := json.Marshal(map[string][]string{"urls": urls})
	if err != nil {
		return nil, err
	}

	// add timeout to the context
	ctx, cancel := context.WithTimeout(ctx, viper.GetDuration("websearch.tavily.timeout"))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", t.BaseURL, bytes.NewBuffer(requestBody))
	t.logger.Debug("Request body", zap.String("body", string(requestBody)))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	t.logger.Debug("Response status", zap.Int("status", resp.StatusCode))
	t.logger.Debug("Response headers", zap.Any("headers", resp.Header))

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result models.WebPages
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
