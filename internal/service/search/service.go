package search

import (
	"context"
	"fmt"

	"github.com/dimdasci/seek/internal/client/openai"
	"go.uber.org/zap"
)

type Service struct {
	openaiClient *openai.Client
	logger       *zap.Logger
}

func NewService(openaiClient *openai.Client, logger *zap.Logger) *Service {
	return &Service{
		openaiClient: openaiClient,
		logger:       logger,
	}
}

func (s *Service) Search(ctx context.Context, query string) (string, error) {
	s.logger.Info("Service: searching for answer",
		zap.String("query", query))

	searchPlan, err := s.openaiClient.PlanSearch(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to search answer: %w", err)
	}

	return searchPlan.String(), nil
}
