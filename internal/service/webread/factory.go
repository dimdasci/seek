package webread

import (
	"context"
	"sync"
	"time"

	"github.com/dimdasci/seek/internal/models"
	"go.uber.org/zap"
)

// ReaderFactory creates the appropriate WebReader based on the URL
type ReaderFactory struct {
	logger         *zap.Logger
	timeout        time.Duration
	minContentLen  int
	browserReader  *BrowserReadService
	standardReader *ReadService
}

func NewReaderFactory(logger *zap.Logger, timeout time.Duration, minContentLen int) (*ReaderFactory, error) {
	browserReader, err := NewBrowserReadService(logger, timeout)
	if err != nil {
		return nil, err
	}

	return &ReaderFactory{
		logger:         logger,
		timeout:        timeout,
		minContentLen:  minContentLen,
		browserReader:  browserReader,
		standardReader: NewReadService(logger, timeout),
	}, nil
}

func (f *ReaderFactory) GetReader() WebReader {
	// Create a composite reader that tries standard first, falls back to browser
	return &fallbackReader{
		primary:       f.standardReader,
		fallback:      f.browserReader,
		logger:        f.logger,
		minContentLen: f.minContentLen,
	}
}

type fallbackReader struct {
	primary       WebReader
	fallback      WebReader
	logger        *zap.Logger
	minContentLen int
}

func (f *fallbackReader) Read(ctx context.Context, urls []string) (*models.WebPages, error) {
	var webPages models.WebPages
	var wg sync.WaitGroup
	results := make(chan models.Page, len(urls))
	errors := make(chan models.PageError, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Try primary reader first
			result, err := f.primary.Read(ctx, []string{url})
			contentLength := 0
			if result != nil && len(result.Pages) == 1 {
				contentLength = len(result.Pages[0].Content)
			}

			if err != nil || result == nil || contentLength < f.minContentLen {
				f.logger.Info("Primary reader failed for URL, trying fallback",
					zap.String("url", url),
					zap.Error(err),
					zap.Int("pages", func() int {
						if result != nil {
							return len(result.Pages)
						}
						return 0
					}()),
					zap.Int("content_length", contentLength),
					zap.Int("min_content_length", f.minContentLen))

				// Try fallback reader
				result, err = f.fallback.Read(ctx, []string{url})
				if err != nil || result == nil || len(result.Pages) != 1 {
					f.logger.Error("Both readers failed", zap.String("url", url), zap.Error(err))
					errors <- models.PageError{URL: url, Error: err.Error()}
					return
				}
				f.logger.Info("Fallback reader succeeded", zap.String("url", url),
					zap.Int("pages", len(result.Pages)),
					zap.Int("page length", len(result.Pages[0].Content)))
			} else {
				f.logger.Info("Primary reader succeeded", zap.String("url", url),
					zap.Int("pages", len(result.Pages)),
					zap.Int("page length", contentLength),
					zap.Int("min_content_length", f.minContentLen))
			}

			if len(result.Pages) > 0 {
				results <- result.Pages[0]
			}
			if len(result.Errors) > 0 {
				errors <- result.Errors[0]
			}
		}(url)
	}

	wg.Wait()
	close(results)
	close(errors)

	for result := range results {
		webPages.Pages = append(webPages.Pages, result)
	}
	for err := range errors {
		webPages.Errors = append(webPages.Errors, err)
	}

	return &webPages, nil
}

func (f *ReaderFactory) Close() error {
	return f.browserReader.Close()
}
