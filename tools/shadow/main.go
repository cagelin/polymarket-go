// Command shadow checks declarations without the exceptions in govet's shadow pass.
package main

import (
	"fmt"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var analyzer = &analysis.Analyzer{
	Name: "noshadow",
	Doc:  "reject declarations that hide enclosing declarations or predeclared identifiers",
	Run:  run,
}

func main() { singlechecker.Main(analyzer) }

type finding struct {
	position token.Pos
	name     string
	outer    types.Object
}

func findShadows(info *types.Info) []finding {
	seen := make(map[token.Pos]bool)
	var findings []finding
	check := func(object types.Object) {
		if object == nil || object.Name() == "_" || object.Parent() == nil || seen[object.Pos()] {
			return
		}
		if variable, ok := object.(*types.Var); ok && variable.IsField() {
			return // Fields belong to a type, not a lexical scope.
		}
		parent := object.Parent().Parent()
		if parent == nil {
			return
		}
		_, outer := parent.LookupParent(object.Name(), object.Pos())
		if outer != nil {
			seen[object.Pos()] = true
			findings = append(findings, finding{position: object.Pos(), name: object.Name(), outer: outer})
		}
	}
	for _, object := range info.Defs {
		check(object)
	}
	for _, object := range info.Implicits {
		check(object) // Type-switch case variables and implicit import names.
	}
	slices.SortFunc(findings, func(a, b finding) int {
		if a.position < b.position {
			return -1
		}
		if a.position > b.position {
			return 1
		}
		return 0
	})
	return findings
}

func run(pass *analysis.Pass) (any, error) {
	for _, item := range findShadows(pass.TypesInfo) {
		origin := "Go predeclared identifier"
		if item.outer.Parent() != types.Universe {
			origin = fmt.Sprintf("declaration at %s", pass.Fset.Position(item.outer.Pos()))
		}
		pass.Reportf(item.position, "declaration of %q shadows %s", item.name, origin)
	}
	return nil, nil
}
