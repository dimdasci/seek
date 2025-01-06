package search

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (s *Service) executeSimpleSearch(ctx context.Context, query string, policy string) string {
	fmt.Printf("Query:       %s\n", query)

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

	fmt.Printf("Read  %d pages, got %d error\n", len(pages.Pages), len(pages.Errors))

	// call compiler
	answer := s.openaiClient.CompileResults(ctx, pages.Pages, &query, &policy)

	// build map of URLs to page titles from search results
	// titles := make(map[string]string, len(results))
	// for _, result := range results {
	// 	titles[result.URL] = result.Title
	// }

	// var notes string
	// for _, page := range pages.Pages {
	// 	notes += fmt.Sprintf("## %s\n\n%s\n\n%s\n\n", titles[page.URL], page.URL, page.Content)
	// }
	return answer //fmt.Sprintf("%v\n\n%v", answer, notes)
}
