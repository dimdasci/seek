package webread

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/dimdasci/seek/internal/models"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

// BrowserReadService provides web reading functionality for SPA and JavaScript-rendered pages.
type BrowserReadService struct {
	logger       *zap.Logger
	timeout      time.Duration
	cache        sync.Map
	tagsToRemove map[string]struct{}
	browser      *rod.Browser
}

func NewBrowserReadService(logger *zap.Logger, timeout time.Duration) (*BrowserReadService, error) {
	// Launch a new browser
	url := launcher.New().
		Headless(true).
		MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()

	return &BrowserReadService{
		logger: logger,
		tagsToRemove: map[string]struct{}{
			"header": {},
			"footer": {},
			"nav":    {},
			"aside":  {},
			"script": {},
			"head":   {},
			"form":   {},
			"select": {},
		},
		timeout: timeout,
		browser: browser,
	}, nil
}

func (b *BrowserReadService) Close() error {
	return b.browser.Close()
}

func (b *BrowserReadService) Read(ctx context.Context, urls []string) (*models.WebPages, error) {
	var wg sync.WaitGroup
	results := make(chan models.Page, len(urls))
	errors := make(chan models.PageError, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			if cached, ok := b.cache.Load(url); ok {
				b.logger.Debug("Returning cached result", zap.String("url", url))
				results <- cached.(models.Page)
				return
			}

			// Create a new page
			page := b.browser.MustPage()
			defer page.Close()

			// Set timeout for navigation
			page.Timeout(b.timeout)

			// Navigate to the URL
			if err := page.Navigate(url); err != nil {
				b.logger.Error("Failed to navigate to URL", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}

			// Wait for the page to be loaded
			if err := page.WaitLoad(); err != nil {
				b.logger.Error("Failed to wait for page load", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}

			// Wait for the page to become visually stable
			if err := page.WaitStable(b.timeout); err != nil {
				b.logger.Error("Failed to wait for page to stabilize", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}

			// Get the page title
			var title string
			titleObj, err := page.Eval(`() => document.title`)
			if err != nil {
				b.logger.Error("Failed to get page title", zap.String("url", url), zap.Error(err))
				title = "Untitled"
			} else {
				title = titleObj.Value.String()
			}

			// Get the page HTML
			rawHTML, err := page.HTML()
			if err != nil {
				b.logger.Error("Failed to get page HTML", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}

			// Parse and clean HTML
			doc, err := html.Parse(strings.NewReader(rawHTML))
			if err != nil {
				b.logger.Error("Failed to parse HTML content", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}

			doc = b.removeUnwantedTags(doc)

			var buf bytes.Buffer
			if err := html.Render(&buf, doc); err != nil {
				b.logger.Error("Failed to render cleaned HTML", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}
			cleanedHTML := buf.String()

			// Convert HTML to markdown
			markdown, err := htmltomarkdown.ConvertString(cleanedHTML)

			if err != nil {
				b.logger.Error("Failed to convert HTML to markdown", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}

			result := models.Page{
				URL:     url,
				Title:   title,
				Content: markdown,
			}

			results <- result
			b.cache.Store(url, result)
		}(url)
	}

	wg.Wait()
	close(results)
	close(errors)

	var webPages models.WebPages
	for result := range results {
		webPages.Pages = append(webPages.Pages, result)
	}
	for err := range errors {
		webPages.Errors = append(webPages.Errors, err)
	}

	return &webPages, nil
}

// removeUnwantedTags removes unwanted tags from an HTML node and returns the cleaned node.
func (b *BrowserReadService) removeUnwantedTags(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}

	// If the current node is an unwanted tag, skip it and its children
	if n.Type == html.ElementNode {
		if _, found := b.tagsToRemove[n.Data]; found {
			return nil
		}
	}

	// Create a copy of the current node
	newNode := &html.Node{
		Type:     n.Type,
		DataAtom: n.DataAtom,
		Data:     n.Data,
		Attr:     append([]html.Attribute(nil), n.Attr...),
	}

	// Recursively process the children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		newChild := b.removeUnwantedTags(c)
		if newChild != nil {
			if newNode.FirstChild == nil {
				newNode.FirstChild = newChild
			} else {
				newNode.LastChild.NextSibling = newChild
				newChild.PrevSibling = newNode.LastChild
			}
			newNode.LastChild = newChild
		}
	}

	return newNode
}
