package search

import (
	"context"
	"fmt"

	"github.com/dimdasci/seek/internal/models"
	"go.uber.org/zap"
)

// executeComplexSearch performs a complex search for the given steps.
// It returns a string with the search results.
func (s *Service) executeComplexSearch(ctx context.Context, plan *models.Plan) string {
	var topics string = ""
	var outline string = ""

	for i, step := range plan.SearchPlan {
		policy := fmt.Sprintf("%s\n\n%s", step.SubRequest, step.FinalAnswerOutline)

		outline += fmt.Sprintf("- %d. %s\n", i+1, step.Topic)

		fmt.Printf(
			"Step %d. %s\n", i+1, step.Topic)

		s.logger.Debug("Complex search step",
			zap.Int("step", i+1),
			zap.String("topic", step.Topic),
			zap.String("sub_request", step.SubRequest),
			zap.String("search_query", step.SearchQuery),
			zap.String("final_answer_outline", step.FinalAnswerOutline))

		switch step.SearchQuery {
		case "":
			if topics == "" {
				s.logger.Debug("Topics are empty for an empty search query",
					zap.Int("step", i+1),
					zap.String("topic", step.Topic))
				continue
			}
			topics += "\n\n" + s.openaiClient.CompileFindings(topics, step.Topic, policy)
		default:
			topics += "\n\n" + s.executeSimpleSearch(ctx,
				step.Topic,
				step.SearchQuery,
				policy)
		}

		s.logger.Debug("Complex search step result",
			zap.Int("step", i+1),
			zap.String("topic", step.Topic))

	}
	fmt.Print("Working on the final answer...\n\n")
	return s.openaiClient.WriteReport(
		ctx,
		&topics,
		&plan.SearchQuery,
		&outline,
		&plan.CompilationPolicy)
}
