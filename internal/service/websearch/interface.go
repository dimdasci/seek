// Package websearch provides a service for web search.
package websearch

import (
	"context"

	"github.com/dimdasci/seek/internal/models"
)

// SearchOptions defines parameters for customizing the search
type SearchOptions struct {
	// DateRestrict limits results to specified timeframe
	// Examples: "d1" (past 24h), "w2" (past 2 weeks), "m6" (past 6 months), "y1" (past year)
	DateRestrict string

	// Language restricts results to a specific language (e.g., "lang_en", "lang_fr")
	Language string

	// Country restricts results to a specific country (e.g., "countryUS", "countryGB")
	Country string

	// Site restricts results to specific domain (e.g., "example.com")
	Site string

	// ExcludeSites excludes results from specific domains
	ExcludeSite string

	// SafeSearch enables/disables safe search filtering ("active" or "off")
	SafeSearch string

	// ResultsPerPage specifies number of results to return (1-10)
	ResultsPerPage int
}

// WebSearcher defines the interface for web search functionality.
type WebSearcher interface {
	Search(ctx context.Context, query string, opts *SearchOptions) ([]models.SearchResult, error)
}
