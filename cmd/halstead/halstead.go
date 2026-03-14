package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	halstead "github.com/luisantonioig/halstead-metrics"
)

func main() {
	jsonOutput := flag.Bool("json", false, "emit machine-readable JSON output")
	changedOnly := flag.Bool("changed-only", false, "when comparing against a baseline, report only changed or new functions")
	maxVolume := flag.Float64("max-volume", 0, "fail if file or function volume exceeds this value")
	maxDifficulty := flag.Float64("max-difficulty", 0, "fail if file or function difficulty exceeds this value")
	maxEffort := flag.Float64("max-effort", 0, "fail if file or function effort exceeds this value")
	baselineReport := flag.String("baseline-report", "", "compare current analysis against a saved JSON report")
	baselineGit := flag.String("baseline-git", "", "compare against the file version stored at the given git revision")
	maxVolumeDelta := flag.Float64("max-volume-delta", 0, "fail if file or function volume increases more than this amount from baseline")
	maxDifficultyDelta := flag.Float64("max-difficulty-delta", 0, "fail if file or function difficulty increases more than this amount from baseline")
	maxEffortDelta := flag.Float64("max-effort-delta", 0, "fail if file or function effort increases more than this amount from baseline")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: halstead [--json] [--changed-only] [--max-volume N] [--max-difficulty N] [--max-effort N] [--baseline-report report.json] [--baseline-git REV] [--max-volume-delta N] [--max-difficulty-delta N] [--max-effort-delta N] <file.go>")
		os.Exit(2)
	}
	if *baselineReport != "" && *baselineGit != "" {
		fmt.Fprintln(os.Stderr, "use only one of --baseline-report or --baseline-git")
		os.Exit(2)
	}

	path := flag.Arg(0)
	src, err := os.ReadFile(path)
	check(err)

	report, err := halstead.AnalyzeASTReport(src)
	check(err)
	report.Path = path
	thresholds := halstead.Thresholds{
		MaxVolume:     *maxVolume,
		MaxDifficulty: *maxDifficulty,
		MaxEffort:     *maxEffort,
	}
	deltaThresholds := halstead.DeltaThresholds{
		MaxVolumeDelta:     *maxVolumeDelta,
		MaxDifficultyDelta: *maxDifficultyDelta,
		MaxEffortDelta:     *maxEffortDelta,
	}
	applyThresholds(&report, thresholds)
	if *baselineReport != "" || *baselineGit != "" {
		baseline, baselineLabel, err := loadBaseline(path, *baselineReport, *baselineGit)
		check(err)
		comparison := halstead.CompareReports(baseline, report)
		comparison.BaselinePath = baselineLabel
		if *changedOnly {
			comparison = comparison.ChangedOnly()
			filterReportFunctionsToComparison(&report, comparison)
		}
		comparison.Evaluate(deltaThresholds)
		report.Comparison = &comparison
	}
	if *jsonOutput {
		printJSON(report)
		if reportFailed(report) || comparisonFailed(report) {
			os.Exit(1)
		}
		return
	}
	printMetrics(halstead.Metrics{
		Name:           report.File.Name,
		Operators:      report.File.Operators,
		Operands:       report.File.Operands,
		TotalOperators: report.File.TotalOperators,
		TotalOperands:  report.File.TotalOperands,
	})
	printFunctions(report.Functions)
	printThresholdSummary(report)
	printComparisonSummary(report)
	if reportFailed(report) || comparisonFailed(report) {
		os.Exit(1)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func printMetrics(metrics halstead.Metrics) {
	fmt.Printf("Analyzer: %s\n", metrics.Name)
	fmt.Println("Operators")
	fmt.Println("--------------------------------")
	printCounts(metrics.Operators)
	fmt.Println()
	fmt.Println("Operands")
	fmt.Println("--------------------------------")
	printCounts(metrics.Operands)
	fmt.Println()
	fmt.Printf("Existen %d operadores diferentes\n", metrics.DifferentOperators())
	fmt.Printf("Existen %d operandos diferentes\n", metrics.DifferentOperands())
	fmt.Printf("El codigo tiene %d operandos y %d operadores\n", metrics.TotalOperands, metrics.TotalOperators)
	fmt.Printf("El tamanio calculado del programa es %f y el volumen es %f\n", metrics.CalculatedProgramLength(), metrics.Volume())
	fmt.Printf("La dificultad del programa es %f\n", metrics.Difficulty())
	fmt.Printf("El esfuerzo del programa es %f\n", metrics.Effort())
	fmt.Printf("El tiempo requerido para programar es %f\n", metrics.TimeRequiredToProgram())
	fmt.Printf("El numero de bugs es %f\n", metrics.NumberOfDeliveredBugs())
}

func printFunctions(functions []halstead.FunctionReport) {
	if len(functions) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Functions")
	fmt.Println("--------------------------------")
	for _, function := range functions {
		fmt.Printf("%s [%s] %d:%d-%d:%d\n", function.Name, function.Kind, function.Start.Line, function.Start.Column, function.End.Line, function.End.Column)
		fmt.Printf("  volume=%f difficulty=%f effort=%f\n", function.Metrics.Volume, function.Metrics.Difficulty, function.Metrics.Effort)
		if function.Threshold != nil && !function.Threshold.Passed {
			for _, violation := range function.Threshold.Violations {
				fmt.Printf("  threshold failed: %s actual=%f limit=%f\n", violation.Metric, violation.Actual, violation.Limit)
			}
		}
	}
}

func printJSON(report halstead.AnalysisReport) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	check(encoder.Encode(report))
}

func loadBaselineReport(path string) (halstead.AnalysisReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return halstead.AnalysisReport{}, err
	}
	var report halstead.AnalysisReport
	if err := json.Unmarshal(data, &report); err != nil {
		return halstead.AnalysisReport{}, err
	}
	return report, nil
}

func loadBaseline(currentPath, baselineReportPath, baselineGitRev string) (halstead.AnalysisReport, string, error) {
	if baselineReportPath != "" {
		report, err := loadBaselineReport(baselineReportPath)
		if err != nil {
			return halstead.AnalysisReport{}, "", err
		}
		return report, baselineReportPath, nil
	}

	report, err := loadBaselineFromGit(currentPath, baselineGitRev)
	if err != nil {
		return halstead.AnalysisReport{}, "", err
	}
	return report, report.Path, nil
}

func loadBaselineFromGit(currentPath, rev string) (halstead.AnalysisReport, error) {
	absPath, err := filepath.Abs(currentPath)
	if err != nil {
		return halstead.AnalysisReport{}, err
	}
	repoRoot, err := gitOutput(filepath.Dir(absPath), "rev-parse", "--show-toplevel")
	if err != nil {
		return halstead.AnalysisReport{}, fmt.Errorf("resolve git repo root: %w", err)
	}
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return halstead.AnalysisReport{}, err
	}
	if strings.HasPrefix(relPath, "..") {
		return halstead.AnalysisReport{}, fmt.Errorf("file %s is outside git repo %s", absPath, repoRoot)
	}
	relPath = filepath.ToSlash(relPath)

	data, err := gitOutput(filepath.Dir(absPath), "show", rev+":"+relPath)
	if err != nil {
		return halstead.AnalysisReport{}, fmt.Errorf("load baseline from git %s:%s: %w", rev, relPath, err)
	}
	report, err := halstead.AnalyzeASTReport([]byte(data))
	if err != nil {
		return halstead.AnalysisReport{}, err
	}
	report.Path = "git:" + rev + ":" + relPath
	return report, nil
}

func gitOutput(workdir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func filterReportFunctionsToComparison(report *halstead.AnalysisReport, comparison halstead.ComparisonOutcome) {
	allowed := make(map[string]struct{}, len(comparison.Functions))
	for _, function := range comparison.Functions {
		allowed[comparisonKey(function.Name, function.Kind)] = struct{}{}
	}

	filtered := make([]halstead.FunctionReport, 0, len(report.Functions))
	for _, function := range report.Functions {
		if _, ok := allowed[comparisonKey(function.Name, function.Kind)]; ok {
			filtered = append(filtered, function)
		}
	}
	report.Functions = filtered
}

func printCounts(values map[string]int) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Println("Key:", key, "Value:", values[key])
	}
}

func applyThresholds(report *halstead.AnalysisReport, thresholds halstead.Thresholds) {
	if !thresholds.Enabled() {
		return
	}
	report.Threshold = report.File.Evaluate(thresholds)
	for i := range report.Functions {
		report.Functions[i].Threshold = report.Functions[i].Metrics.Evaluate(thresholds)
	}
}

func printThresholdSummary(report halstead.AnalysisReport) {
	if report.Threshold == nil {
		return
	}
	fmt.Println()
	fmt.Println("Thresholds")
	fmt.Println("--------------------------------")
	if report.Threshold.Passed {
		fmt.Println("File metrics passed configured thresholds.")
	} else {
		fmt.Println("File metrics exceeded configured thresholds:")
		for _, violation := range report.Threshold.Violations {
			fmt.Printf("  %s actual=%f limit=%f\n", violation.Metric, violation.Actual, violation.Limit)
		}
	}
}

func reportFailed(report halstead.AnalysisReport) bool {
	if report.Threshold != nil && !report.Threshold.Passed {
		return true
	}
	for _, function := range report.Functions {
		if function.Threshold != nil && !function.Threshold.Passed {
			return true
		}
	}
	return false
}

func printComparisonSummary(report halstead.AnalysisReport) {
	if report.Comparison == nil {
		return
	}
	fmt.Println()
	fmt.Println("Comparison")
	fmt.Println("--------------------------------")
	fmt.Printf("Baseline report: %s\n", report.Comparison.BaselinePath)
	fmt.Printf("File delta: volume=%f difficulty=%f effort=%f\n", report.Comparison.File.Delta.VolumeDelta, report.Comparison.File.Delta.DifficultyDelta, report.Comparison.File.Delta.EffortDelta)
	if report.Comparison.File.Threshold != nil && !report.Comparison.File.Threshold.Passed {
		for _, violation := range report.Comparison.File.Threshold.Violations {
			fmt.Printf("  file delta threshold failed: %s delta=%f limit=%f\n", violation.Metric, violation.Delta, violation.Limit)
		}
	}
	for _, function := range report.Comparison.Functions {
		fmt.Printf("%s [%s] delta: volume=%f difficulty=%f effort=%f\n", function.Name, function.Kind, function.Delta.VolumeDelta, function.Delta.DifficultyDelta, function.Delta.EffortDelta)
		if !function.FoundInBase {
			fmt.Println("  new function in current report")
		}
		if function.Threshold != nil && !function.Threshold.Passed {
			for _, violation := range function.Threshold.Violations {
				fmt.Printf("  function delta threshold failed: %s delta=%f limit=%f\n", violation.Metric, violation.Delta, violation.Limit)
			}
		}
	}
}

func comparisonFailed(report halstead.AnalysisReport) bool {
	if report.Comparison == nil || report.Comparison.Threshold == nil {
		return false
	}
	if !report.Comparison.Threshold.Passed {
		return true
	}
	if report.Comparison.File.Threshold != nil && !report.Comparison.File.Threshold.Passed {
		return true
	}
	for _, function := range report.Comparison.Functions {
		if function.Threshold != nil && !function.Threshold.Passed {
			return true
		}
	}
	return false
}

func comparisonKey(name, kind string) string {
	return kind + ":" + name
}
