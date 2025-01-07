package models

import "fmt"

// SearchResult represents a single search result from Tavily API.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// String returns the string representation of the search result.
func (r *SearchResult) String() string {
	return fmt.Sprintf("## %s\n\nURL: %s\n\n%s\n\n", r.Title, r.URL, r.Content)
}
