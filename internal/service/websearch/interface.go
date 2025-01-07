// Package websearch provides a service for web search.
package websearch

import (
	"context"

	"github.com/dimdasci/seek/internal/models"
)

// WebSearcher defines the interface for web search functionality.
type WebSearcher interface {
	Search(ctx context.Context, query string) ([]models.SearchResult, error)
}
