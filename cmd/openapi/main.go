// Command openapi derives the committed specification from route and wire models.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/paths"
)

func main() {
	check := flag.Bool("check", false, "verify committed artifact is current")

	flag.Parse()

	schemas := api.NewSchemaRegistry()
	document := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Organization Purchasing Example",
			"version":     "0.1.0",
			"description": "Local ergonomics prototype; demo bearer auth only. Business validation remains in operations.",
		},
		"servers": []any{map[string]any{"url": "http://127.0.0.1:8080"}},
		"components": map[string]any{
			"securitySchemes": map[string]any{"demoBearer": map[string]any{
				"type":   "http",
				"scheme": "bearer",
			}},
			"schemas": schemas.Components,
			"responses": map[string]any{
				"Error": response(
					"Error envelope; HTTP status identifies invalid input, auth, policy, conflict or internal failure.",
					schemas.Reference(api.ErrorResponse{}),
				),
			},
		},
	}
	routes := map[string]any{}

	for _, endpoint := range paths.All() {
		parameters := []any{}
		if endpoint.Authenticated {
			parameters = append(parameters, map[string]any{
				"name":        "X-Organization-ID",
				"in":          "header",
				"required":    true,
				"description": "Active organization, verified against token membership.",
				"schema":      map[string]any{"type": "string"},
			})
		}

		for _, part := range strings.Split(endpoint.Path, "/") {
			if strings.HasPrefix(part, "{") {
				parameters = append(parameters, map[string]any{
					"name":     strings.Trim(part, "{}"),
					"in":       "path",
					"required": true,
					"schema":   map[string]any{"type": "string"},
				})
			}
		}

		operation := map[string]any{
			"operationId":  endpoint.ID,
			"summary":      endpoint.Name,
			"description":  endpoint.Description,
			"parameters":   parameters,
			"x-permission": endpoint.Permission,
			"x-feature":    endpoint.Feature,
		}
		if endpoint.Authenticated {
			operation["security"] = []any{
				map[string]any{"demoBearer": []string{}},
			}
		}

		responses := map[string]any{
			strconv.Itoa(endpoint.SuccessStatus): response(
				"Success",
				schemas.Reference(endpoint.Response),
			),
		}
		responses["default"] = map[string]any{
			"$ref": "#/components/responses/Error",
		}

		operation["responses"] = responses
		if endpoint.Request != nil {
			operation["requestBody"] = map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": schemas.Reference(endpoint.Request),
					},
				},
			}
		}

		pathItem, ok := routes[endpoint.Path].(map[string]any)
		if !ok {
			pathItem = map[string]any{}
			routes[endpoint.Path] = pathItem
		}

		pathItem[strings.ToLower(endpoint.Method)] = operation
	}

	document["paths"] = routes

	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		panic(err)
	}

	encoded = append(encoded, '\n')

	const destination = "openapi/openapi.json"
	if *check {
		existing, err := os.ReadFile(destination)
		if err != nil || string(existing) != string(encoded) {
			fmt.Fprintln(os.Stderr, "OpenAPI differs; run go run ./cmd/openapi")
			os.Exit(1)
		}

		fmt.Println("OpenAPI artifact is current")

		return
	}

	if err := os.WriteFile(destination, encoded, 0o644); err != nil {
		panic(err)
	}
}

func response(description string, schema map[string]any) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{"schema": schema},
		},
	}
}
