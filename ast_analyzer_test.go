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

func TestAnalyzeASTReportIncludesFunctions(t *testing.T) {
	src := []byte(`package main

import "fmt"

func helper(v int) int {
	return v + 1
}

func main() {
	fmt.Println(helper(2))
}
`)

	report, err := AnalyzeASTReport(src)
	if err != nil {
		t.Fatalf("AnalyzeASTReport returned error: %v", err)
	}

	if report.Analyzer != "go-ast-types" {
		t.Fatalf("unexpected analyzer: %q", report.Analyzer)
	}
	if len(report.Functions) != 2 {
		t.Fatalf("expected 2 function reports, got %d", len(report.Functions))
	}
	if report.Functions[0].Name != "helper" {
		t.Fatalf("expected first function to be helper, got %q", report.Functions[0].Name)
	}
	if report.Functions[1].Name != "main" {
		t.Fatalf("expected second function to be main, got %q", report.Functions[1].Name)
	}
	if report.Functions[0].Metrics.TotalOperators == 0 || report.Functions[1].Metrics.TotalOperands == 0 {
		t.Fatalf("expected non-empty per-function metrics, got %+v", report.Functions)
	}
}

func TestThresholdEvaluation(t *testing.T) {
	summary := MetricsSummary{
		Volume:     20,
		Difficulty: 5,
		Effort:     100,
	}

	outcome := summary.Evaluate(Thresholds{
		MaxVolume:     10,
		MaxDifficulty: 6,
		MaxEffort:     50,
	})

	if outcome == nil {
		t.Fatalf("expected threshold outcome")
	}
	if outcome.Passed {
		t.Fatalf("expected threshold evaluation to fail")
	}
	if len(outcome.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(outcome.Violations))
	}
}

func TestCompareReports(t *testing.T) {
	baseline := AnalysisReport{
		Path: "/tmp/base.json",
		File: MetricsSummary{Volume: 10, Difficulty: 2, Effort: 20},
		Functions: []FunctionReport{
			{Name: "main", Kind: "func_decl", Metrics: MetricsSummary{Volume: 5, Difficulty: 1, Effort: 5}},
		},
	}
	current := AnalysisReport{
		File: MetricsSummary{Volume: 20, Difficulty: 4, Effort: 50},
		Functions: []FunctionReport{
			{Name: "main", Kind: "func_decl", Metrics: MetricsSummary{Volume: 9, Difficulty: 2, Effort: 12}},
			{Name: "helper", Kind: "func_decl", Metrics: MetricsSummary{Volume: 8, Difficulty: 2, Effort: 16}},
		},
	}

	comparison := CompareReports(baseline, current)
	if comparison.File.Delta.VolumeDelta != 10 {
		t.Fatalf("expected file volume delta 10, got %f", comparison.File.Delta.VolumeDelta)
	}
	if len(comparison.Functions) != 2 {
		t.Fatalf("expected 2 function comparisons, got %d", len(comparison.Functions))
	}
	if comparison.Functions[1].FoundInBase {
		t.Fatalf("expected helper not to be found in baseline")
	}
	if !comparison.Functions[1].IsChanged {
		t.Fatalf("expected helper to be marked as changed")
	}

	comparison.Evaluate(DeltaThresholds{MaxVolumeDelta: 6})
	if comparison.Threshold == nil || comparison.Threshold.Passed {
		t.Fatalf("expected comparison thresholds to fail")
	}
}

func TestCompareReportsChangedOnly(t *testing.T) {
	baseline := AnalysisReport{
		Functions: []FunctionReport{
			{Name: "same", Kind: "func_decl", Metrics: MetricsSummary{Volume: 5, Difficulty: 1, Effort: 5}},
			{Name: "changed", Kind: "func_decl", Metrics: MetricsSummary{Volume: 5, Difficulty: 1, Effort: 5}},
		},
	}
	current := AnalysisReport{
		Functions: []FunctionReport{
			{Name: "same", Kind: "func_decl", Metrics: MetricsSummary{Volume: 5, Difficulty: 1, Effort: 5}},
			{Name: "changed", Kind: "func_decl", Metrics: MetricsSummary{Volume: 7, Difficulty: 1, Effort: 8}},
			{Name: "newfn", Kind: "func_decl", Metrics: MetricsSummary{Volume: 3, Difficulty: 1, Effort: 3}},
		},
	}

	filtered := CompareReports(baseline, current).ChangedOnly()
	if len(filtered.Functions) != 2 {
		t.Fatalf("expected 2 changed functions, got %d", len(filtered.Functions))
	}
	if filtered.Functions[0].Name != "changed" || filtered.Functions[1].Name != "newfn" {
		t.Fatalf("unexpected changed-only functions: %+v", filtered.Functions)
	}
}
