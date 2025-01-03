/* Package plan provides a model for a plan. */
package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// Plan represents a search plan.
type Plan struct {
	Approved          bool     `json:"approved"`           // Approved field set to false if request contains illegal content or other instructions
	Reason            string   `json:"reason"`             // Reason for approval or rejection
	SearchQuery       *string  `json:"search_query"`       // Web search query for simple requests, null for complex requests
	SearchComplexity  string   `json:"search_complexity"`  // Search complexity: simple or complex
	SearchPlan        []Search `json:"search_plan"`        // Search plan for the request
	CompilationPolicy string   `json:"compilation_policy"` // Policy to compile the findings into the final report
}

// Search represents a search plan for a specific topic.
type Search struct {
	Topic              string `json:"topic"`                // Topic of the search
	SearchQuery        string `json:"search_query"`         // Web search query
	SubRequest         string `json:"sub_request"`          // Sub-request to conduct the information gathering
	FinalAnswerOutline string `json:"final_answer_outline"` // Outline of the final answer
}

var (
	// Match content between ```json and ``` markers
	jsonBlockRegex = regexp.MustCompile("(?s)```json\\s*\\n(.*?)\\n\\s*```")
)

// NewPlan creates a new search plan from stringified JSON.
func NewPlan(jsonStr string) (*Plan, error) {
	var err error
	plan := &Plan{}

	// extract JSON content from markdown code blocks
	if strings.Contains(jsonStr, "```json") {
		if jsonStr, err = extractJSONFromMarkdown(jsonStr); err != nil {
			return nil, err
		}
	}
	if err = json.Unmarshal([]byte(jsonStr), plan); err != nil {
		return nil, err
	}

	if err = plan.validate(); err != nil {
		return nil, err
	}

	return plan, nil
}

// String returns the string representation of the search plan.
func (p *Plan) String() string {
	b, _ := json.MarshalIndent(p, "", "  ")
	return string(b)
}

// validate ensures the plan is valid according to business rules
func (p *Plan) validate() error {
	// Must have a reason if not approved
	if !p.Approved && p.Reason == "" {
		return errors.New("reason is required when plan is not approved")
	}

	// Validate based on complexity
	switch p.SearchComplexity {
	case "simple":
		if p.SearchQuery == nil {
			return errors.New("searchQuery is required for simple searches")
		}
		if len(p.SearchPlan) > 0 {
			return errors.New("search_plan should be empty for simple searches")
		}
	case "complex":
		if p.SearchQuery != nil {
			return errors.New("searchQuery should be null for complex searches")
		}
		if len(p.SearchPlan) == 0 {
			return errors.New("search_plan is required for complex searches")
		}
	default:
		return errors.New("invalid search complexity")
	}

	return nil
}

// extractJSONFromMarkdown extracts JSON content from markdown code blocks
func extractJSONFromMarkdown(markdown string) (string, error) {
	// Find all JSON blocks
	matches := jsonBlockRegex.FindStringSubmatch(markdown)
	if len(matches) < 2 {
		return "", errors.New("no JSON code block found in markdown")
	}

	// Return the content of the first JSON block (matches[1] contains the capture group)
	return strings.TrimSpace(matches[1]), nil
}
