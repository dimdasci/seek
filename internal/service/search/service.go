package search

import (
	"context"
	"fmt"

	"github.com/dimdasci/seek/internal/client/openai"
	"github.com/dimdasci/seek/internal/models"
	"github.com/dimdasci/seek/internal/service/websearch"
	"go.uber.org/zap"
)

type Service struct {
	openaiClient *openai.Client
	searcher     websearch.WebSearcher
	logger       *zap.Logger
}

func NewService(openaiClient *openai.Client, searcher websearch.WebSearcher, logger *zap.Logger) *Service {
	return &Service{
		openaiClient: openaiClient,
		searcher:     searcher,
		logger:       logger,
	}
}

func (s *Service) Search(ctx context.Context, query string) (string, error) {
	s.logger.Info("Service: searching for answer",
		zap.String("query", query))

	fmt.Println("Building search plan...")
	p, err := s.openaiClient.PlanSearch(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to search answer: %w", err)
	}

	if p == nil {
		return "", fmt.Errorf("search plan is nil")
	}

	if !p.Approved {
		return "", fmt.Errorf("search plan is not approved: %s", p.Reason)
	}

	fmt.Printf("Going to perform %s search\n", p.SearchComplexity)

	report, err := s.executePlan(ctx, p)
	if err != nil {
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
	answer, _, err := s.searcher.Search(ctx, query)
	if err != nil {
		return fmt.Sprintf("Failed to search for %s", query)
	}

	// join results into a single string
	// var notes string
	// for _, result := range results {
	// 	notes += result.String()
	// }
	return answer
}

func (s *Service) executeComplexSearch(ctx context.Context, steps []models.Search) string {
	fmt.Printf("Steps:\n")
	for i, step := range steps {
		fmt.Printf("%d. %s [%s]\n", i+1, step.SubRequest, step.SearchQuery)
	}
	return "Complex search executed"
}
