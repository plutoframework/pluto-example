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
			"description": "Local prototype using Pluto sessions; demo credentials only bootstrap sessions. Business validation remains in operations.",
		},
		"servers": []any{map[string]any{"url": "http://127.0.0.1:8080"}},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"sessionCookie": map[string]any{
					"type": "apiKey",
					"in":   "cookie",
					"name": "__Host-pluto_session",
				},
				"bearerAuth": map[string]any{
					"type":        "http",
					"scheme":      "bearer",
					"description": "Opaque Pluto session token; never the demo bootstrap credential.",
				},
				"demoCredential": map[string]any{
					"type":        "http",
					"scheme":      "bearer",
					"description": "Local fixture credential, accepted only by POST /auth/session.",
				},
			},
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
		if endpoint.Authenticated && !endpoint.SessionOnly {
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

		if endpoint.Collection {
			parameters = append(
				parameters,
				map[string]any{
					"name": "limit",
					"in":   "query",
					"schema": map[string]any{
						"type":    "integer",
						"minimum": 1,
						"maximum": 100,
						"default": 25,
					},
				},
				map[string]any{
					"name": "cursor",
					"in":   "query",
					"schema": map[string]any{
						"type":      "string",
						"maxLength": 2048,
					},
				},
			)
		}

		if endpoint.Idempotent {
			parameters = append(
				parameters,
				map[string]any{
					"name":        "Idempotency-Key",
					"in":          "header",
					"required":    true,
					"description": "Reuse for retries of this action. 409 means conflict or pending reconciliation. Completed retention 24h.",
					"schema": map[string]any{
						"type":      "string",
						"minLength": 1,
						"maxLength": 128,
						"pattern":   "^[A-Za-z0-9_.:-]+$",
					},
				},
			)
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
				map[string]any{"bearerAuth": []string{}},
				map[string]any{"sessionCookie": []string{}},
			}
		}

		if endpoint.ID == paths.CreateSession.ID {
			operation["security"] = []any{
				map[string]any{"demoCredential": []string{}},
			}

			parameters = append(
				parameters,
				map[string]any{
					"name":     "Origin",
					"in":       "header",
					"required": true,
					"schema":   map[string]any{"type": "string"},
				},
			)
			operation["parameters"] = parameters
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

	// Keep the spec and schemas readable and individually below 1,000 lines.
	references := map[string]any{}
	for name := range schemas.Components {
		references[name] = map[string]any{
			"$ref": "./schemas.json#/components/schemas/" + name,
		}
	}

	document["components"].(map[string]any)["schemas"] = references

	artifacts := map[string]any{
		"openapi/openapi.json": document,
		"openapi/schemas.json": map[string]any{
			"components": map[string]any{"schemas": schemas.Components},
		},
	}
	for destination, value := range artifacts {
		encoded, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			panic(err)
		}

		encoded = append(encoded, '\n')
		if *check {
			existing, err := os.ReadFile(destination)
			if err != nil || string(existing) != string(encoded) {
				fmt.Fprintln(
					os.Stderr,
					"OpenAPI differs; run go run ./cmd/openapi",
				)
				os.Exit(1)
			}
		} else if err := os.WriteFile(destination, encoded, 0o644); err != nil {
			panic(err)
		}
	}

	if *check {
		fmt.Println("OpenAPI artifacts are current")
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
