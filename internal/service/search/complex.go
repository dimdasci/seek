package search

import (
	"context"
	"fmt"

	"github.com/dimdasci/seek/internal/models"
)

func (s *Service) executeComplexSearch(ctx context.Context, steps []models.Search) string {
	fmt.Printf("Steps:\n")
	for i, step := range steps {
		fmt.Printf("%d. %s [%s]\n", i+1, step.SubRequest, step.SearchQuery)
	}
	return "Complex search executed"
}
