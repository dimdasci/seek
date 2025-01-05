package search

import (
	"context"
	"fmt"

	"github.com/dimdasci/seek/internal/client/openai"
	"github.com/dimdasci/seek/internal/models"
	"github.com/dimdasci/seek/internal/service/webread"
	"github.com/dimdasci/seek/internal/service/websearch"
	"go.uber.org/zap"
)

type Service struct {
	openaiClient *openai.Client
	searcher     websearch.WebSearcher
	reader       webread.WebReader
	logger       *zap.Logger
}

func NewService(
	openaiClient *openai.Client,
	searcher websearch.WebSearcher,
	reader webread.WebReader,
	logger *zap.Logger) *Service {
	return &Service{
		openaiClient: openaiClient,
		searcher:     searcher,
		reader:       reader,
		logger:       logger,
	}
}

func (s *Service) Search(ctx context.Context, query string) (string, error) {
	s.logger.Info("Service: searching for answer",
		zap.String("query", query))

	fmt.Println("Building search plan...")
	p, err := s.openaiClient.PlanSearch(ctx, query)
	if err != nil {
		s.logger.Error("Service: failed to search for answer",
			zap.Error(err))
		return "", fmt.Errorf("failed to search answer: %w", err)
	}

	if p == nil {
		s.logger.Error("Service: search plan is nil")
		return "", fmt.Errorf("search plan is nil")
	}

	if !p.Approved {
		s.logger.Error("Service: search plan is not approved",
			zap.String("reason", p.Reason))
		return "", fmt.Errorf("search plan is not approved: %s", p.Reason)
	}

	fmt.Printf("Going to perform %s search\n", p.SearchComplexity)

	report, err := s.executePlan(ctx, p)
	if err != nil {
		s.logger.Error("Service: failed to execute search plan",
			zap.Error(err))
		return "", fmt.Errorf("failed to execute search plan: %w", err)
	}
	return report, nil
}

func (s *Service) executePlan(ctx context.Context, plan *models.Plan) (string, error) {
	if plan == nil {
		return "", fmt.Errorf("search plan is nil")
	}

	// execute simple search
	var notes string
	switch plan.SearchComplexity {
	case "simple":
		notes = s.executeSimpleSearch(ctx, *plan.SearchQuery)
	case "complex":
		notes = s.executeComplexSearch(ctx, plan.SearchPlan)
	default:
		return "", fmt.Errorf("unknown search complexity: %s", plan.SearchComplexity)
	}

	return notes, nil
}

func (s *Service) executeSimpleSearch(ctx context.Context, query string) string {
	fmt.Printf("Query:       %s\n", query)

	// perform web search
	answer, results, err := s.searcher.Search(ctx, query)
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

	// build map of URLs to page titles from search results
	titles := make(map[string]string, len(results))
	for _, result := range results {
		titles[result.URL] = result.Title
	}

	var notes string
	for _, page := range pages.Pages {
		notes += fmt.Sprintf("## %s\n\n%s\n\n%s\n\n", titles[page.URL], page.URL, page.Content)
	}
	return fmt.Sprintf("%v\n\n%v", answer, notes)
}

func (s *Service) executeComplexSearch(ctx context.Context, steps []models.Search) string {
	fmt.Printf("Steps:\n")
	for i, step := range steps {
		fmt.Printf("%d. %s [%s]\n", i+1, step.SubRequest, step.SearchQuery)
	}
	return "Complex search executed"
}
