package openai

import (
	"github.com/invopop/jsonschema"
)

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
