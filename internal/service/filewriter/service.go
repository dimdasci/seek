package filewriter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dimdasci/seek/internal/models"
	"go.uber.org/zap"
)

// Service handles saving web pages to files
type Service struct {
	logger    *zap.Logger
	outputDir string
}

// Add as package-level variable
var (
	invalidChars = regexp.MustCompile(`[^a-z0-9-_]`)
	multipleDash = regexp.MustCompile(`-+`)
)

func NewService(logger *zap.Logger, outputDir string) (*Service, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &Service{
		logger:    logger,
		outputDir: outputDir,
	}, nil
}

func (s *Service) SavePages(pages *models.WebPages) {
	for _, page := range pages.Pages {
		if err := s.SavePage(page); err != nil {
			s.logger.Error("Failed to save page",
				zap.String("url", page.URL),
				zap.Error(err))
		}
	}

	if len(pages.Errors) > 0 {
		fmt.Println("\nErrors occurred while processing the following URLs:")
		for _, err := range pages.Errors {
			fmt.Printf("- %s: %s\n", err.URL, err.Error)
		}
	}
}

func (s *Service) SavePage(page models.Page) error {
	filename := s.generateFilename(page.Title, page.URL)
	filepath := filepath.Join(s.outputDir, filename)

	if err := os.WriteFile(filepath, []byte(page.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}

	s.logger.Info("Saved page content",
		zap.String("url", page.URL),
		zap.String("filepath", filepath))
	fmt.Printf("Saved content from %s to %s\n", page.URL, filepath)
	return nil
}

func (s *Service) generateFilename(title, url string) string {
	name := title
	if name == "" {
		urlParts := strings.Split(strings.TrimRight(url, "/"), "/")
		for i := len(urlParts) - 1; i >= 0; i-- {
			if part := strings.TrimSpace(urlParts[i]); part != "" {
				name = part
				break
			}
		}
		if name == "" {
			name = "untitled"
		}
	}

	// Clean and normalize
	name = strings.ToLower(name)
	name = invalidChars.ReplaceAllString(name, "-")
	name = multipleDash.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")

	if len(name) > 100 {
		name = name[:100]
	}

	return name + ".md"
}
