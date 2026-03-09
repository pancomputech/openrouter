package openrouter

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

// SchemaFor generates a JSON Schema for T using invopop/jsonschema.
func SchemaFor[T any]() json.RawMessage {
	r := &jsonschema.Reflector{}
	var zero T
	schema := r.Reflect(zero)
	b, _ := json.Marshal(schema)
	return b
}
