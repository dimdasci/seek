package search

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// executeSimpleSearch performs a simple search for the given query.
// It returns a string with the search results.
func (s *Service) executeSimpleSearch(
	ctx context.Context,
	topic string,
	query string,
	policy string,
) string {
	fmt.Printf("Query: %s\n\n", query)

	s.logger.Debug("Simple search",
		zap.String("query", query),
		zap.String("policy", policy))

	// perform web search
	results, err := s.searcher.Search(ctx, query)
	if err != nil {
		s.logger.Error("Service: failed to search for answer",
			zap.Error(err))
		return fmt.Sprintf("Failed to search for %s", query)
	}

	urls := make([]string, 0, len(results))
	for _, result := range results {
		urls = append(urls, result.URL)
	}

	s.logger.Info("Service: read web pages",
		zap.Int("pages", len(urls)))

	pages, err := s.reader.Read(ctx, urls)
	if err != nil {
		s.logger.Error("Service: failed to read web pages",
			zap.Error(err))
		return fmt.Sprintf("Failed to read web pages: %v", err)
	}

	s.logger.Debug("Service: read web pages",
		zap.Int("pages", len(pages.Pages)),
		zap.Int("errors", len(pages.Errors)))

	for _, page := range pages.Errors {
		s.logger.Error("Service: failed to read web page",
			zap.String("url", page.URL),
			zap.String("error", page.Error))
	}

	answer := s.openaiClient.CompileResults(ctx, pages.Pages, &topic, &policy)

	return answer
}
