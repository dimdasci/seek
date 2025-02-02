package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dimdasci/seek/internal/search"
)

// List of supported language codes by Google CSE API
var googleSupportedLanguages = map[string]bool{
	"ar": true, "bg": true, "ca": true, "cs": true, "da": true,
	"de": true, "el": true, "en": true, "es": true, "et": true,
	"fi": true, "fr": true, "hr": true, "hu": true, "id": true,
	"is": true, "it": true, "iw": true, "ja": true, "ko": true,
	"lt": true, "lv": true, "nl": true, "no": true, "pl": true,
	"pt": true, "ro": true, "ru": true, "sk": true, "sl": true,
	"sr": true, "sv": true, "tr": true,
	"zh-CN": true, "zh-TW": true, // Special cases for Chinese
}

// GoogleEngine implements the search.Engine interface for Google Custom Search
type GoogleEngine struct {
	apiKey   string
	cx       string
	endPoint string
	client   *http.Client
}

// googleSearchResponse represents the JSON response from Google CSE API
type googleSearchResponse struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
}

// NewGoogleEngine creates a new Google Custom Search engine instance
func NewGoogleEngine(apiKey, cx, endPoint string) (*GoogleEngine, error) {
	if apiKey == "" || cx == "" || endPoint == "" {
		return nil, fmt.Errorf("apiKey, cx, and endPoint are required")
	}

	return &GoogleEngine{
		apiKey:   apiKey,
		cx:       cx,
		endPoint: endPoint,
		client:   &http.Client{},
	}, nil
}

// Name returns the name of the search engine
func (e *GoogleEngine) Name() string {
	return "google"
}

// Search performs a search using Google Custom Search API
func (e *GoogleEngine) Search(ctx context.Context, criteria *search.Criteria) ([]search.Result, error) {
	// Build search URL with query parameters
	params := url.Values{}
	params.Add("key", e.apiKey)
	params.Add("cx", e.cx)
	params.Add("q", criteria.Query)

	// Add optional parameters
	if criteria.MaxResults > 0 {
		params.Add("num", fmt.Sprintf("%d", criteria.MaxResults))
	}
	if criteria.Language != "" {
		// Validate language code for Google CSE
		if !googleSupportedLanguages[criteria.Language] {
			return nil, fmt.Errorf("unsupported language code %q for Google search. Supported codes: ar,bg,ca,cs,da,de,el,en,es,et,fi,fr,hr,hu,id,is,it,iw,ja,ko,lt,lv,nl,no,pl,pt,ro,ru,sk,sl,sr,sv,tr,zh-CN,zh-TW", criteria.Language)
		}
		params.Add("lr", fmt.Sprintf("lang_%s", criteria.Language))
	}
	if criteria.DateRestrict != "" {
		params.Add("dateRestrict", criteria.DateRestrict)
	}
	if criteria.IncludeDomain != "" {
		params.Add("siteSearch", criteria.IncludeDomain)
		params.Add("siteSearchFilter", "i")
	}
	if criteria.ExcludeDomain != "" {
		params.Add("siteSearch", criteria.ExcludeDomain)
		params.Add("siteSearchFilter", "e")
	}
	if criteria.SafeSearch {
		params.Add("safe", "active")
	}

	// Timeout
	ctx, cancel := context.WithTimeout(ctx, criteria.TimeoutDuration)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s?%s", e.endPoint, params.Encode()),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var googleResp googleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Google results to our Result type
	results := make([]search.Result, 0, len(googleResp.Items))
	t := time.Now()
	for _, item := range googleResp.Items {
		results = append(results, search.Result{
			Title:     item.Title,
			URL:       item.Link,
			Snippet:   item.Snippet,
			Source:    e.Name(),
			FetchedAt: t,
		})
	}

	return results, nil
}
