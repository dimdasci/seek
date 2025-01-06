package openai

import (
	"github.com/invopop/jsonschema"
)

// Structured output schemas
type AnalysisResult struct {
	ReasoningSteps []Step `json:"reasoning_steps" jsonschema_description:"The chain of reasoning"`
	Relevance      bool   `json:"relevance" jsonschema_description:"The relevance of the page to the request"`
	Answer         string `json:"answer" jsonschema_description:"The key points from the page"`
}

type CompilationResult struct {
	ReasoningSteps []Step `json:"reasoning_steps" jsonschema_description:"The chain of reasoning"`
	Compilation    string `json:"compilation" jsonschema_description:"The compiled result"`
}

type Step struct {
	Explanation string `json:"explanation"`
	Output      string `json:"output"`
}

func GenerateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}
