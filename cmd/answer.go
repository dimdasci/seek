/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dimdasci/seek/internal/client/openai"
	"github.com/dimdasci/seek/internal/service/search"
	"github.com/dimdasci/seek/internal/service/webread"
	"github.com/dimdasci/seek/internal/service/websearch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	// Set default values for timeouts
	viper.SetDefault("openai.reasoning.timeout", "60s")
	viper.SetDefault("websearch.tavily.timeout", "10s")
	viper.SetDefault("webread.timeout", "10s")

	answerCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for the answer (markdown format)")
}

func runAnswerCmd(cmd *cobra.Command, args []string) {
	// Combine all args into the question
	question := strings.Join(args, " ")
	logger.Info("Searching for an answer", zap.String("question", question))

	apiKey := viper.GetString("openai.api_key")
	if apiKey == "" {
		logger.Error("OpenAI API key not found. Exiting...")
		fmt.Println("OpenAI API key not found. Exiting...")
		return
	}
	// Initialize clients and services
	openaiClient := openai.NewClient(apiKey, logger)
	webSearcher := websearch.NewTavilySearchService(logger, viper.GetDuration("websearch.tavily.timeout"))
	webReader := webread.NewReadService(logger, viper.GetDuration("webread.timeout"))
	searchService := search.NewService(openaiClient, webSearcher, webReader, logger)

	// Context with timeout
	ctx := context.Background()

	// Search for the answer
	answer, err := searchService.Search(ctx, question)
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
