package api

import (
	"reflect"
	"strings"
	"time"
)

// SchemaRegistry is intentionally small: it supports the concrete wire models used
// in this example, not arbitrary Go types or a general validation DSL.
type SchemaRegistry struct {
	Components map[string]any
}

func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{Components: make(map[string]any)}
}

func (r *SchemaRegistry) Reference(value any) map[string]any {
	return r.schema(reflect.TypeOf(value))
}

func (r *SchemaRegistry) schema(value reflect.Type) map[string]any {
	if value == reflect.TypeOf(time.Time{}) {
		return map[string]any{
			"type":   "string",
			"format": "date-time",
		}
	}

	switch value.Kind() {
	case reflect.Struct:
		if _, exists := r.Components[value.Name()]; exists {
			return map[string]any{
				"$ref": "#/components/schemas/" + value.Name(),
			}
		}

		r.Components[value.Name()] = nil
		properties := map[string]any{}
		required := []string{}

		for index := 0; index < value.NumField(); index++ {
			field := value.Field(index)

			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				continue
			}

			properties[name] = r.schema(field.Type)
			required = append(required, name)
		}

		r.Components[value.Name()] = map[string]any{
			"type":                 "object",
			"properties":           properties,
			"required":             required,
			"additionalProperties": false,
		}

		return map[string]any{"$ref": "#/components/schemas/" + value.Name()}
	case reflect.Slice:
		return map[string]any{
			"type":  "array",
			"items": r.schema(value.Elem()),
		}
	case reflect.Map:
		return map[string]any{
			"type":                 "object",
			"additionalProperties": r.schema(value.Elem()),
		}
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int64:
		return map[string]any{
			"type":   "integer",
			"format": "int64",
		}
	default:
		panic("unsupported wire schema type: " + value.String())
	}
}
