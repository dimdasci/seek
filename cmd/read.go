package cmd

import (
	"context"
	"fmt"

	"github.com/dimdasci/seek/internal/config"
	"github.com/dimdasci/seek/internal/service/filewriter"
	"github.com/dimdasci/seek/internal/service/webread"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var outputDir string

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read [urls]",
	Short: "Read the content of the given URLs and convert it to markdown",
	Long: `Read command fetches content from provided URLs, converts it to markdown format,
and saves each page to a separate file. The filename is generated from the page title.`,
	Args: cobra.MinimumNArgs(1),
	Run:  runReadCmd,
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "directory to save markdown files")
}

func runReadCmd(cmd *cobra.Command, args []string) {
	cfg := config.Get()

	writer, err := filewriter.NewService(logger, outputDir)
	if err != nil {
		logger.Error("Failed to initialize file writer", zap.Error(err))
		fmt.Printf("Failed to initialize file writer: %v\n", err)
		return
	}

	logger.Debug("Initializing web reader", zap.Duration("timeout", cfg.WebReader.Timeout), zap.Int("min_content_length", cfg.WebReader.MinContentLength))
	readerFactory, err := webread.NewReaderFactory(logger, cfg.WebReader.Timeout, cfg.WebReader.MinContentLength)
	if err != nil {
		logger.Error("Failed to initialize web reader", zap.Error(err))
		fmt.Printf("Failed to initialize web reader: %v\n", err)
		return
	}
	defer readerFactory.Close()

	reader := readerFactory.GetReader()
	webPages, err := reader.Read(context.Background(), args)
	if err != nil {
		logger.Error("Failed to read web pages", zap.Error(err))
		fmt.Printf("Failed to read web pages: %v\n", err)
		return
	}

	writer.SavePages(webPages)
}
