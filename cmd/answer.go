package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dimdasci/seek/internal/client/openai"
	"github.com/dimdasci/seek/internal/config"
	"github.com/dimdasci/seek/internal/service/search"
	"github.com/dimdasci/seek/internal/service/webread"
	"github.com/dimdasci/seek/internal/service/websearch"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var outputFile string

// answerCmd represents the answer command
var answerCmd = &cobra.Command{
	Use:   "answer [question]",
	Short: "Search for an answer to your question",
	Args:  cobra.MinimumNArgs(1),
	Run:   runAnswerCmd,
}

func init() {
	rootCmd.AddCommand(answerCmd)

	answerCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for the result in markdown format")
}

func runAnswerCmd(cmd *cobra.Command, args []string) {
	question := strings.Join(args, " ")
	logger.Info("Searching for an answer", zap.String("question", question))

	cfg := config.Get()
	if cfg.OpenAI.APIKey == "" {
		logger.Error("OpenAI API key not found. Exiting...")
		fmt.Println("OpenAI API key not found. Exiting...")
		return
	}

	// Initialize clients and services
	openaiClient, err := openai.NewClient(
		cfg.OpenAI.APIKey,
		logger,
		cfg.OpenAI.Reasoning.Model,
		cfg.OpenAI.Completion.Model,
		cfg.OpenAI.Reasoning.Timeout,
		cfg.OpenAI.Completion.Timeout,
		cfg.OpenAI.Reasoning.MaxTokens,
		cfg.OpenAI.Completion.MaxTokens,
	)
	if err != nil {
		logger.Error("Failed to create OpenAI client", zap.Error(err))
		fmt.Printf("Failed to create OpenAI client: %v\n", err)
		return
	}
	webSearcher := websearch.NewTavilySearchService(logger, cfg.WebSearch.Tavily.Timeout)
	webReader := webread.NewReadService(logger, cfg.WebRead.Timeout)
	searchService := search.NewService(openaiClient, webSearcher, webReader, logger)

	// Search for the answer
	answer, err := searchService.Search(context.Background(), question)
	if err != nil {
		logger.Error("Failed to get answer", zap.Error(err))
		fmt.Printf("Failed to get answer: %v\n", err)
		return
	}

	// Handle output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(answer), 0644); err != nil {
			logger.Error("Failed to write to file", zap.Error(err))
			fmt.Printf("failed to write to file: %v", err)
		}
		fmt.Printf("Answer saved to: %s\n", outputFile)
	} else {
		fmt.Println(answer)
	}
	logger.Info("Answer found", zap.String("answer", answer))
}
