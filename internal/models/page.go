package models

// Page represents a web page content.
type Page struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Content string `json:"raw_content"`
}

// PageError represents a web page error.
type PageError struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

// WebPages represents a collection of web pages.
type WebPages struct {
	Pages  []Page      `json:"results"`
	Errors []PageError `json:"failed_results"`
}
