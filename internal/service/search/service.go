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
		notes = s.executeSimpleSearch(ctx, *plan.SearchQuery, plan.CompilationPolicy)
	case "complex":
		notes = s.executeComplexSearch(ctx, plan.SearchPlan)
	default:
		return "", fmt.Errorf("unknown search complexity: %s", plan.SearchComplexity)
	}

	return notes, nil
}
