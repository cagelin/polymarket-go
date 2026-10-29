// Command exportinventory lists type and constant declarations for the facade generator.
// Input directories must be platform independent: this tool inventories every non-test Go file
// without filtering build tags or platform suffixes.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type declaration struct {
	Kind       string
	Name       string
	Directory  string
	Parameters string
	Arguments  string
}

func inspect(source []byte, directory string) ([]declaration, error) {
	fileset := token.NewFileSet()
	parsed, err := parser.ParseFile(fileset, "source.go", source, 0)
	if err != nil {
		return nil, err
	}
	var declarations []declaration
	for _, node := range parsed.Decls {
		general, ok := node.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range general.Specs {
			switch item := spec.(type) {
			case *ast.TypeSpec:
				if !item.Name.IsExported() {
					continue
				}
				entry := declaration{Kind: "type", Name: item.Name.Name, Directory: directory}
				if item.TypeParams != nil {
					var parameters, arguments []string
					for _, field := range item.TypeParams.List {
						// Fail explicitly on constraints needing extra imports instead of dropping an API.
						if _, simple := field.Type.(*ast.Ident); !simple {
							return nil, fmt.Errorf("%s: configure facade imports for complex generic constraint", entry.Name)
						}
						var constraint bytes.Buffer
						if err = format.Node(&constraint, fileset, field.Type); err != nil {
							return nil, err
						}
						text := constraint.String()
						if ast.IsExported(text) {
							text = filepath.Base(directory) + "." + text
						}
						for _, name := range field.Names {
							parameters = append(parameters, name.Name+" "+text)
							arguments = append(arguments, name.Name)
						}
					}
					entry.Parameters = "[" + strings.Join(parameters, ", ") + "]"
					entry.Arguments = "[" + strings.Join(arguments, ", ") + "]"
				}
				declarations = append(declarations, entry)
			case *ast.ValueSpec:
				// Exported variables are mutable state, not aliases; intentionally exclude them.
				if general.Tok != token.CONST {
					continue
				}
				for _, name := range item.Names {
					if name.IsExported() {
						declarations = append(declarations, declaration{Kind: "const", Name: name.Name, Directory: directory})
					}
				}
			}
		}
	}
	return declarations, nil
}
func run() error {
	var declarations []declaration
	for _, directory := range os.Args[1:] {
		paths, err := filepath.Glob(filepath.Join(directory, "*.go"))
		if err != nil {
			return err
		}
		for _, sourcePath := range paths {
			if strings.HasSuffix(sourcePath, "_test.go") {
				continue
			}
			//nolint:gosec // Paths are developer-selected Go source files for AST inspection.
			source, readErr := os.ReadFile(sourcePath)
			if readErr != nil {
				return readErr
			}
			entries, parseErr := inspect(source, directory)
			if parseErr != nil {
				return fmt.Errorf("%s: %w", sourcePath, parseErr)
			}
			declarations = append(declarations, entries...)
		}
	}
	return json.NewEncoder(os.Stdout).Encode(declarations)
}
func main() {
	if err := run(); err != nil {
		// A failed diagnostic write cannot be recovered; preserve the failure exit status.
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
