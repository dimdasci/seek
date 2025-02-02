package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/dimdasci/seek/internal/logging"
	"github.com/dimdasci/seek/internal/search"
)

var (
	maxResults    int
	searchEngine  string
	language      string
	dateRestrict  string
	includeDomain string
	excludeDomain string
	safeSearch    bool
	timeout       time.Duration
)

func init() {
	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search the web for information",
		Long: `Search command performs a web search using the specified search engine 
and returns results matching the query. For example:

seek search "golang concurrency patterns"
seek search --engine google --max-results 5 "kubernetes best practices"
seek search --include-domain github.com "go testing examples"`,
		Args: cobra.MinimumNArgs(1),
		RunE: runSearch,
	}

	// Add flags
	searchCmd.Flags().IntVarP(&maxResults, "max-results", "n", 5, "maximum number of results to return")
	searchCmd.Flags().StringVarP(&searchEngine, "engine", "e", "google", "search engine to use (google, bing)")
	searchCmd.Flags().StringVarP(&language, "language", "l", "", "limit results to specific language (e.g., en, es)")
	searchCmd.Flags().StringVar(&dateRestrict, "date", "", "limit by date (d1, w1, m1, y1 for day/week/month/year)")
	searchCmd.Flags().StringVarP(&includeDomain, "include-domain", "i", "", "limit results to specific domain")
	searchCmd.Flags().StringVarP(&excludeDomain, "exclude-domain", "x", "", "exclude results from specific domain")
	searchCmd.Flags().BoolVar(&safeSearch, "safe-search", false, "enable/disable safe search")

	// Add command to root
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()

	// Combine all arguments into the search query
	query := args[0]
	if len(args) > 1 {
		for _, arg := range args[1:] {
			query += " " + arg
		}
	}

	// Create search criteria from flags
	criteria := search.NewCriteria(query).
		WithMaxResults(maxResults).
		WithLanguage(language).
		WithDateRestrict(dateRestrict).
		WithIncludeDomain(includeDomain).
		WithExcludeDomain(excludeDomain).
		WithSafeSearch(safeSearch).
		WithTimeout(viper.GetDuration("search.timeout"))

	// Log the search parameters
	logger.Info("Starting search",
		zap.String("query", criteria.Query),
		zap.String("engine", searchEngine),
		zap.Int("max_results", criteria.MaxResults),
		zap.String("language", criteria.Language),
		zap.String("date_restrict", criteria.DateRestrict),
		zap.String("include_domain", criteria.IncludeDomain),
		zap.String("exclude_domain", criteria.ExcludeDomain),
		zap.Bool("safe_search", criteria.SafeSearch),
		zap.Duration("timeout", criteria.TimeoutDuration),
	)

	// Create context with timeout
	_, cancel := context.WithTimeout(cmd.Context(), criteria.TimeoutDuration)
	defer cancel()

	// For now, just print a message
	cmd.Printf("Searching for: <%s> using %s engine (max results: %d)\n",
		query, searchEngine, maxResults)

	return nil
}
