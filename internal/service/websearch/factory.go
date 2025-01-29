package websearch

import (
	"fmt"

	"github.com/dimdasci/seek/internal/config"
	"go.uber.org/zap"
)

// SearcherFactory creates WebSearcher instances
type SearcherFactory struct {
	logger *zap.Logger
}

// NewSearcherFactory creates a new SearcherFactory
func NewSearcherFactory(logger *zap.Logger) *SearcherFactory {
	return &SearcherFactory{
		logger: logger,
	}
}

// Create returns a WebSearcher implementation based on the given name
func (f *SearcherFactory) Create(name string, cfg *config.Config) (WebSearcher, error) {
	switch name {
	case "google":
		return NewGoogleSearch(f.logger, cfg), nil
	case "tavily":
		return NewTavilySearch(f.logger, cfg), nil
	default:
		return nil, fmt.Errorf("unknown search engine: %s", name)
	}
}
