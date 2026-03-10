package openrouter

import (
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
)

// SchemaFor generates a JSON Schema for T using invopop/jsonschema.
// It returns the flat object schema suitable for use as tool parameters.
func SchemaFor[T any]() json.RawMessage {
	r := &jsonschema.Reflector{}
	var zero T
	doc := r.Reflect(zero)
	// Named structs are placed in $defs with a $ref at the root.
	// Extract the definition directly so callers get type:"object" at the top level.
	if name := reflect.TypeOf(zero).Name(); name != "" {
		if def, ok := doc.Definitions[name]; ok {
			b, _ := json.Marshal(def)
			return b
		}
	}
	b, _ := json.Marshal(doc)
	return b
}
