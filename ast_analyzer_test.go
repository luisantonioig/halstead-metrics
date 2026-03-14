package halstead

import "testing"

func TestAnalyzeASTBasicProgram(t *testing.T) {
	src := []byte(`package main

import "fmt"

func main() {
	x := 1 + 2
	fmt.Println(x)
}`)

	metrics, err := AnalyzeAST(src)
	if err != nil {
		t.Fatalf("AnalyzeAST returned error: %v", err)
	}

	if metrics.Name != "go-ast-types" {
		t.Fatalf("unexpected analyzer name: %q", metrics.Name)
	}
	if metrics.Operators["func"] != 1 {
		t.Fatalf("expected func operator to be counted once, got %d", metrics.Operators["func"])
	}
	if metrics.Operators[":="] != 1 {
		t.Fatalf("expected := operator to be counted once, got %d", metrics.Operators[":="])
	}
	if metrics.Operators["+"] != 1 {
		t.Fatalf("expected + operator to be counted once, got %d", metrics.Operators["+"])
	}
	if metrics.Operators["call"] != 1 {
		t.Fatalf("expected call operator to be counted once, got %d", metrics.Operators["call"])
	}
	if metrics.Operators["import"] != 0 {
		t.Fatalf("expected import not to be counted as operator, got %d", metrics.Operators["import"])
	}
	if metrics.Operands["var:x"] == 0 {
		t.Fatalf("expected operand x to be counted")
	}
	if metrics.Operands["1"] != 1 || metrics.Operands["2"] != 1 {
		t.Fatalf("expected literals 1 and 2 to be counted once, got 1=%d 2=%d", metrics.Operands["1"], metrics.Operands["2"])
	}
	if metrics.Operands["pkg:fmt"] == 0 {
		t.Fatalf("expected package operand fmt to be counted")
	}
	if metrics.Operands["func:Println"] == 0 {
		t.Fatalf("expected function operand Println to be counted")
	}
}

func TestAnalyzeASTKeywordPolicy(t *testing.T) {
	src := []byte(`package main

import "fmt"

const answer = 42

type message string

var greeting message = "hola"

func main() {
	fmt.Println(greeting, answer)
}`)

	metrics, err := AnalyzeAST(src)
	if err != nil {
		t.Fatalf("AnalyzeAST returned error: %v", err)
	}

	if metrics.Operators["const"] != 1 {
		t.Fatalf("expected const operator once, got %d", metrics.Operators["const"])
	}
	if metrics.Operators["type"] != 1 {
		t.Fatalf("expected type operator once, got %d", metrics.Operators["type"])
	}
	if metrics.Operators["var"] != 1 {
		t.Fatalf("expected var operator once, got %d", metrics.Operators["var"])
	}
	if metrics.Operators["package"] != 0 {
		t.Fatalf("expected package not to be counted, got %d", metrics.Operators["package"])
	}
	if metrics.Operators["import"] != 0 {
		t.Fatalf("expected import not to be counted, got %d", metrics.Operators["import"])
	}
}

func TestAnalyzeASTOnFixture(t *testing.T) {
	src := []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello, mundo")
}`)

	astMetrics, err := AnalyzeAST(src)
	if err != nil {
		t.Fatalf("AnalyzeAST returned error: %v", err)
	}

	if astMetrics.TotalOperators == 0 || astMetrics.TotalOperands == 0 {
		t.Fatalf("expected ast analyzer to produce counts, got %+v", astMetrics)
	}
}
