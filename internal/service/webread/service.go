package webread

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/dimdasci/seek/internal/models"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

// ReadService provides web reading functionality.
type ReadService struct {
	logger       *zap.Logger
	tagsToRemove map[string]struct{}
	timeout      time.Duration
}

func NewReadService(logger *zap.Logger, timeout time.Duration) *ReadService {
	return &ReadService{
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
			"image":  {},
		},
		timeout: timeout,
	}
}

func (r *ReadService) Read(ctx context.Context, urls []string) (*models.WebPages, error) {
	var wg sync.WaitGroup
	results := make(chan models.Page, len(urls))
	errors := make(chan models.PageError, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			r.logger.Debug("Reading web page", zap.String("url", url))

			// if the URL is not valid, skip it
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				r.logger.Error("Invalid URL", zap.String("url", url))
				errors <- models.PageError{URL: url, Error: "invalid URL"}
				return
			}

			// if url is PDF, skip it
			if strings.HasSuffix(url, ".pdf") {
				r.logger.Error("PDF file", zap.String("url", url))
				errors <- models.PageError{URL: url, Error: "PDF file"}
				return
			}

			htmlContent, err := r.fetchHTML(ctx, url)
			if err != nil {
				r.logger.Error("Failed to fetch HTML content", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}

			doc, err := html.Parse(strings.NewReader(htmlContent))
			if err != nil {
				r.logger.Error("Failed to parse HTML content", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}

			title, err := r.extractTitle(doc)
			if err != nil {
				r.logger.Error("Failed to extract title", zap.String("url", url), zap.Error(err))
			}

			doc = r.removeUnwantedTags(doc)

			var buf bytes.Buffer
			if err := html.Render(&buf, doc); err != nil {
				r.logger.Error("Failed to render cleaned HTML", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}
			cleanedHTML := buf.String()

			markdown, err := htmltomarkdown.ConvertString(cleanedHTML)

			if err != nil {
				r.logger.Error("Failed to convert HTML to markdown", zap.String("url", url), zap.Error(err))
				errors <- models.PageError{URL: url, Error: err.Error()}
				return
			}
			r.logger.Debug("Converted HTML to markdown", zap.String("url", url), zap.String("title", title))
			results <- models.Page{URL: url, Title: title, Content: markdown}
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

// fetchHTML fetches the HTML content of the given URL.
func (r *ReadService) fetchHTML(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: r.timeout,
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	return buf.String(), nil
}

// extractTitle parses the HTML and extracts the content of the <title> tag.
func (r *ReadService) extractTitle(doc *html.Node) (string, error) {
	var title string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	if title == "" {
		return "", fmt.Errorf("title tag not found")
	}
	return title, nil
}

// removeUnwantedTags removes unwanted tags from an HTML node and returns the cleaned node.
func (r *ReadService) removeUnwantedTags(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}

	// If the current node is an unwanted tag, skip it and its children
	if n.Type == html.ElementNode {
		if _, found := r.tagsToRemove[n.Data]; found {
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
		newChild := r.removeUnwantedTags(c)
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
