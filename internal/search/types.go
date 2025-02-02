package search

import "time"

// Criteria defines the parameters for a search operation
type Criteria struct {
	// Query is the search string
	Query string

	// DateRestrict limits results to a specific time period
	// Examples: "d1" (past 24h), "w1" (past week), "m1" (past month), "y1" (past year)
	DateRestrict string

	// Language restricts results to a specific language (ISO 639-1 code)
	// Example: "en", "es", "fr"
	Language string

	// IncludeDomain limits results to specific domain
	// Example: "github.com"
	IncludeDomain string

	// ExcludeDomain filters out results from specific domain
	// Example: "example.com"
	ExcludeDomain string

	// SafeSearch enables or disables safe search filtering
	SafeSearch bool

	// MaxResults specifies the maximum number of results to return
	MaxResults int

	// TimeoutDuration specifies how long to wait for search results
	TimeoutDuration time.Duration
}

// Result represents a single search result
type Result struct {
	// Title of the search result
	Title string

	// URL of the search result
	URL string

	// Snippet is a brief excerpt or description of the result
	Snippet string

	// Source indicates which search engine provided this result
	Source string

	// FetchedAt records when this result was retrieved
	FetchedAt time.Time
}

// NewCriteria creates a new Criteria with default values
func NewCriteria(query string) *Criteria {
	return &Criteria{
		Query:      query,
		SafeSearch: true,
	}
}

// WithMaxResults sets the maximum number of results to return
func (c *Criteria) WithMaxResults(max int) *Criteria {
	c.MaxResults = max
	return c
}

// WithLanguage sets the language restriction
func (c *Criteria) WithLanguage(lang string) *Criteria {
	c.Language = lang
	return c
}

// WithDateRestrict sets the date restriction
func (c *Criteria) WithDateRestrict(date string) *Criteria {
	c.DateRestrict = date
	return c
}

// WithIncludeDomains sets the domains to include in search results
func (c *Criteria) WithIncludeDomain(domain string) *Criteria {
	c.IncludeDomain = domain
	return c
}

// WithExcludeDomains sets the domains to exclude from search results
func (c *Criteria) WithExcludeDomain(domain string) *Criteria {
	c.ExcludeDomain = domain
	return c
}

// WithSafeSearch sets whether safe search is enabled
func (c *Criteria) WithSafeSearch(enabled bool) *Criteria {
	c.SafeSearch = enabled
	return c
}

// WithTimeout sets the search timeout duration
func (c *Criteria) WithTimeout(d time.Duration) *Criteria {
	c.TimeoutDuration = d
	return c
}
