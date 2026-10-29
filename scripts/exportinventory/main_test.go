package main

import "testing"

func TestGroupedAndGenericDeclarations(t *testing.T) {
	source := []byte("package sample\ntype (First struct{}; Second[T any] []T)\nconst Solo = 1\nconst (One = 1; Two = 2)\nvar Mutable = 3\n")
	entries, err := inspect(source, "models")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("inspect(grouped declarations) count = %d, want 5", len(entries))
	}
	if entries[1].Parameters != "[T any]" || entries[1].Arguments != "[T]" {
		t.Errorf("inspect(generic) = %+v, want parameters and alias arguments", entries[1])
	}
}
