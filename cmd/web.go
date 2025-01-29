/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/dimdasci/seek/internal/config"
	"github.com/dimdasci/seek/internal/service/websearch"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	outputFileWeb string
	searchEngine  string
	dateRestrict  string
	language      string
	country       string
	site          string // single site to include
	excludeSite   string // single site to exclude
	safeSearch    string
	numResults    int
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web [query]",
	Short: "Search the web",
	Args:  cobra.MinimumNArgs(1),
	Run:   runWebCmd,
}

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().StringVarP(&outputFileWeb, "output", "o", "", "output file for the result in markdown format")
	webCmd.Flags().StringVarP(&searchEngine, "engine", "e", "google", "search engine to use (tavily or google)")

	webCmd.Flags().StringVar(&dateRestrict, "date", "", "restrict results by date (e.g., d1, w2, m6, y1)")
	webCmd.Flags().StringVar(&language, "lang", "", "restrict results by language (e.g., lang_en, lang_fr)")
	webCmd.Flags().StringVar(&country, "country", "", "restrict results by country (e.g., countryUS)")
	webCmd.Flags().StringVar(&site, "site", "", "restrict results to specific domain")
	webCmd.Flags().StringVar(&excludeSite, "exclude-site", "", "exclude results from specific domain")
	webCmd.Flags().StringVar(&safeSearch, "safe", "off", "enable/disable safe search (active/off)")
	webCmd.Flags().IntVar(&numResults, "num", 10, "number of results to return (1-10)")
}

func runWebCmd(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")
	logger.Info("Searching the web",
		zap.String("query", query),
		zap.String("engine", searchEngine),
		zap.String("date", dateRestrict),
		zap.String("lang", language),
		zap.String("country", country),
		zap.String("site", site),
		zap.String("exclude-site", excludeSite),
		zap.String("safe", safeSearch),
		zap.Int("num", numResults),
	)

	// Validate search engine choice
	if searchEngine != "tavily" && searchEngine != "google" {
		logger.Error("Invalid search engine specified. Must be 'tavily' or 'google'")
		fmt.Println("Invalid search engine specified. Must be 'tavily' or 'google'")
		return
	}

	cfg := config.Get()
	if searchEngine == "tavily" && cfg.WebSearch.Tavily.APIKey == "" {
		logger.Error("Tavily API key not found. Exiting...")
		fmt.Println("Tavily API key not found. Exiting...")
		return
	}
	if searchEngine == "google" && (cfg.WebSearch.Google.APIKey == "" || cfg.WebSearch.Google.CX == "") {
		logger.Error("Google API key or CX not found. Exiting...")
		fmt.Println("Google API key or CX not found. Exiting...")
		return
	}

	// Create search options
	opts := &websearch.SearchOptions{
		DateRestrict:   dateRestrict,
		Language:       language,
		Country:        country,
		Site:           site,
		ExcludeSite:    excludeSite,
		SafeSearch:     safeSearch,
		ResultsPerPage: numResults,
	}

	searcherFactory := websearch.NewSearcherFactory(logger)
	searcher, err := searcherFactory.Create(searchEngine, cfg)
	if err != nil {
		logger.Error("Failed to create searcher", zap.Error(err))
		fmt.Printf("Failed to create searcher: %v\n", err)
		return
	}

	results, err := searcher.Search(cmd.Context(), query, opts)
	if err != nil {
		logger.Error("Failed to search", zap.Error(err))
		fmt.Printf("Failed to search: %v\n", err)
		return
	}

	// Print results
	for i, result := range results {
		fmt.Printf("\n%d. %s\n%s\n%s\n\n", i+1, result.Title, result.URL, result.Content)
	}
}
