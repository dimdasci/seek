package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dimdasci/seek/internal/config"
	"github.com/dimdasci/seek/internal/models"
	"go.uber.org/zap"
)

type GoogleSearch struct {
	apiKey  string
	cx      string
	logger  *zap.Logger
	BaseURL string
}

func NewGoogleSearch(logger *zap.Logger, cfg *config.Config) *GoogleSearch {
	return &GoogleSearch{
		apiKey:  cfg.WebSearch.Google.APIKey,
		cx:      cfg.WebSearch.Google.CX,
		logger:  logger,
		BaseURL: cfg.WebSearch.Google.SearchURL,
	}
}

type googleSearchResponse struct {
	Items []struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		Snippet     string `json:"snippet"`
		DisplayLink string `json:"displayLink"`
	} `json:"items"`
}

func (g *GoogleSearch) Search(ctx context.Context, query string, opts *SearchOptions) ([]models.SearchResult, error) {
	params := url.Values{}

	// Required parameters
	params.Add("key", g.apiKey)
	params.Add("cx", g.cx)
	params.Add("q", query)

	// Optional parameters based on SearchOptions
	if opts != nil {
		if opts.DateRestrict != "" {
			params.Add("dateRestrict", opts.DateRestrict)
		}
		if opts.Language != "" {
			params.Add("lr", opts.Language)
		}
		if opts.Country != "" {
			params.Add("cr", opts.Country)
		}
		if opts.SafeSearch != "" {
			params.Add("safe", opts.SafeSearch)
		}
		if opts.ResultsPerPage > 0 {
			params.Add("num", fmt.Sprintf("%d", opts.ResultsPerPage))
		}

		// Handle site restrictions - use first site only as Google CSE only supports one
		if opts.Site != "" {
			params.Add("siteSearch", opts.Site)
			params.Add("siteSearchFilter", "i") // include results from this site
		}
		if opts.ExcludeSite != "" {
			params.Add("siteSearch", opts.ExcludeSite)
			params.Add("siteSearchFilter", "e") // exclude results from this site
		}
	}

	// Create request
	reqURL := fmt.Sprintf("%s?%s", g.BaseURL, params.Encode())
	g.logger.Debug("Google CSE request", zap.String("url", reqURL))
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var googleResp googleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Convert to our model
	results := make([]models.SearchResult, 0, len(googleResp.Items))
	for _, item := range googleResp.Items {
		results = append(results, models.SearchResult{
			Title:   item.Title,
			URL:     item.Link,
			Content: item.Snippet,
		})
	}

	return results, nil
}
