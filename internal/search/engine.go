package search

import "context"

// Engine defines the interface for search providers
type Engine interface {
	// Search performs a search using the given criteria and returns results
	Search(ctx context.Context, criteria *Criteria) ([]Result, error)

	// Name returns the name of the search engine
	Name() string
}
