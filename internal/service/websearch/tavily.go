package websearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dimdasci/seek/internal/models" // Adjust the import path accordingly
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// TavilySearchService provides web search functionality.
type TavilySearchService struct {
	APIKey  string
	BaseURL string
	logger  *zap.Logger
	timeout time.Duration
}

// NewTavilySearchService creates a new instance of TavilySearchService.
func NewTavilySearchService(logger *zap.Logger, timeout time.Duration) *TavilySearchService {
	return &TavilySearchService{
		APIKey:  viper.GetString("websearch.tavily.api_key"),
		BaseURL: viper.GetString("websearch.tavily.search_url"),
		logger:  logger,
		timeout: timeout,
	}
}

// TavilyResponse represents the response from Tavily API.
type TavilyResponse struct {
	Answer  string                `json:"answer"`
	Results []models.SearchResult `json:"results"`
}

// Search performs a web search using the Tavily API.
// Returns the answer and search results.
func (s *TavilySearchService) Search(ctx context.Context, query string) (answer string, results []models.SearchResult, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"query":          query,
		"max_results":    viper.GetInt("websearch.tavily.max_results"),
		"include_answer": true,
	})
	s.logger.Debug("Request body", zap.String("body", string(requestBody)))

	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// add timeout to the context
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	s.logger.Debug("Response status", zap.Int("status", resp.StatusCode))
	s.logger.Debug("Response headers", zap.Any("headers", resp.Header))

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var tavilyResp TavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tavilyResp.Answer, tavilyResp.Results, nil
}
