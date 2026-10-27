package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// This deterministic import-graph check enforces technical boundaries. It does
// not semantically prove that a handler contains no business logic.
func TestArchitecture(t *testing.T) {
	root := filepath.Join("..", "..")

	const module = "github.com/plutoframework/pluto-example/"

	graph := map[string][]string{}

	err := filepath.WalkDir(
		root,
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() {
				if entry.Name() == ".git" {
					return filepath.SkipDir
				}

				return nil
			}

			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if strings.Count(string(source), "\n") > 1000 {
				t.Errorf("file exceeds 1,000 lines: %s", path)
			}

			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			relative, err := filepath.Rel(root, filepath.Dir(path))
			if err != nil {
				return err
			}

			packagePath := module + filepath.ToSlash(relative)

			file, err := parser.ParseFile(
				token.NewFileSet(),
				path,
				source,
				parser.ImportsOnly,
			)
			if err != nil {
				return err
			}

			for _, declaration := range file.Imports {
				imported, err := strconv.Unquote(declaration.Path.Value)
				if err != nil {
					return err
				}

				graph[packagePath] = append(graph[packagePath], imported)
			}

			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	for packagePath := range graph {
		forbidden := func(imported string) bool {
			if strings.HasPrefix(packagePath, module+"internal/handlers/") {
				return strings.HasPrefix(
					imported,
					module+"internal/storage/",
				) ||
					strings.HasPrefix(imported, "github.com/jackc/") ||
					imported == "database/sql"
			}

			if packagePath == module+"internal/operations" ||
				packagePath == module+"internal/domain" {
				return imported == "net/http" ||
					strings.HasPrefix(imported, "github.com/go-chi/") ||
					strings.HasPrefix(imported, module+"internal/handlers/") ||
					imported == module+"internal/pluto" ||
					strings.HasPrefix(imported, module+"internal/storage/")
			}

			if packagePath == module+"internal/paths" {
				return strings.HasPrefix(imported, module+"internal/handlers/")
			}

			return false
		}
		visited := map[string]bool{}

		var visit func(string)

		visit = func(current string) {
			if visited[current] {
				return
			}

			visited[current] = true
			for _, imported := range graph[current] {
				if forbidden(imported) {
					t.Errorf(
						"%s reaches forbidden dependency %s",
						packagePath,
						imported,
					)
				}

				if strings.HasPrefix(imported, module) {
					visit(imported)
				}
			}
		}
		visit(packagePath)
	}
}
