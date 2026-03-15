package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	halstead "github.com/luisantonioig/halstead-metrics"
)

const (
	ansiRed   = "\033[31m"
	ansiGreen = "\033[32m"
	ansiBold  = "\033[1m"
	ansiReset = "\033[0m"
)

func main() {
	os.Exit(exitCode())
}

func exitCode() int {
	return runCLI(os.Args[1:])
}

func runCLI(args []string) int {
	return run(args, os.Stdout, os.Stderr)
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("halstead", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: halstead [--json] [--verbose] [--changed-only] [--max-volume N] [--max-difficulty N] [--max-effort N] [--baseline-report report.json] [--baseline-git REV] [--max-volume-delta N] [--max-difficulty-delta N] [--max-effort-delta N] <file.go>")
		flags.PrintDefaults()
	}

	jsonOutput := flags.Bool("json", false, "emit machine-readable JSON output")
	verboseOutput := flags.Bool("verbose", false, "emit detailed operator and operand counts")
	changedOnly := flags.Bool("changed-only", false, "when comparing against a baseline, report only changed or new functions")
	maxVolume := flags.Float64("max-volume", 0, "fail if file or function volume exceeds this value")
	maxDifficulty := flags.Float64("max-difficulty", 0, "fail if file or function difficulty exceeds this value")
	maxEffort := flags.Float64("max-effort", 0, "fail if file or function effort exceeds this value")
	baselineReport := flags.String("baseline-report", "", "compare current analysis against a saved JSON report")
	baselineGit := flags.String("baseline-git", "", "compare against the file version stored at the given git revision")
	maxVolumeDelta := flags.Float64("max-volume-delta", 0, "fail if file or function volume increases more than this amount from baseline")
	maxDifficultyDelta := flags.Float64("max-difficulty-delta", 0, "fail if file or function difficulty increases more than this amount from baseline")
	maxEffortDelta := flags.Float64("max-effort-delta", 0, "fail if file or function effort increases more than this amount from baseline")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() < 1 {
		flags.Usage()
		return 2
	}
	if *baselineReport != "" && *baselineGit != "" {
		fmt.Fprintf(stderr, "halstead: %s use only one of --baseline-report or --baseline-git\n", redLabel("ERROR"))
		return 2
	}

	path := flags.Arg(0)
	report, err := halstead.AnalyzeASTFile(path)
	if err != nil {
		return printError(stderr, err)
	}
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
		if err != nil {
			return printError(stderr, err)
		}
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
		if err := printJSON(stdout, report); err != nil {
			return printError(stderr, err)
		}
		if reportFailed(report) || comparisonFailed(report) {
			return 1
		}
		return 0
	}
	printMetrics(stdout, halstead.Metrics{
		Name:           report.File.Name,
		Operators:      report.File.Operators,
		Operands:       report.File.Operands,
		TotalOperators: report.File.TotalOperators,
		TotalOperands:  report.File.TotalOperands,
	}, *verboseOutput)
	printFunctions(stdout, report.Functions, *verboseOutput)
	printThresholdSummary(stdout, report)
	printComparisonSummary(stdout, report, *verboseOutput)
	if reportFailed(report) || comparisonFailed(report) {
		return 1
	}
	return 0
}

func printError(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "halstead: %s something failed: %v\n", redLabel("ERROR"), err)
	return 1
}

func printMetrics(stdout io.Writer, metrics halstead.Metrics, verbose bool) {
	fmt.Fprintf(stdout, "Analyzer: %s\n", metrics.Name)
	fmt.Fprintf(stdout, "There are %d distinct operators\n", metrics.DifferentOperators())
	fmt.Fprintf(stdout, "There are %d distinct operands\n", metrics.DifferentOperands())
	fmt.Fprintf(stdout, "The code has %d operands and %d operators\n", metrics.TotalOperands, metrics.TotalOperators)
	fmt.Fprintf(stdout, "The calculated program length is %f and the volume is %f\n", metrics.CalculatedProgramLength(), metrics.Volume())
	fmt.Fprintf(stdout, "The program difficulty is %f\n", metrics.Difficulty())
	fmt.Fprintf(stdout, "The program effort is %f\n", metrics.Effort())
	fmt.Fprintf(stdout, "The estimated programming time is %f\n", metrics.TimeRequiredToProgram())
	fmt.Fprintf(stdout, "The estimated number of bugs is %f\n", metrics.NumberOfDeliveredBugs())
	if !verbose {
		return
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Operators")
	fmt.Fprintln(stdout, "--------------------------------")
	printCounts(stdout, metrics.Operators)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Operands")
	fmt.Fprintln(stdout, "--------------------------------")
	printCounts(stdout, metrics.Operands)
}

func printFunctions(stdout io.Writer, functions []halstead.FunctionReport, verbose bool) {
	if len(functions) == 0 {
		return
	}
	if !verbose {
		printFunctionSummary(stdout, functions)
		return
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Functions")
	fmt.Fprintln(stdout, "--------------------------------")
	for _, function := range functions {
		fmt.Fprintf(stdout, "%s [%s] %d:%d-%d:%d\n", function.Name, function.Kind, function.Start.Line, function.Start.Column, function.End.Line, function.End.Column)
		fmt.Fprintf(stdout, "  volume=%f difficulty=%f effort=%f\n", function.Metrics.Volume, function.Metrics.Difficulty, function.Metrics.Effort)
		if function.Threshold != nil && !function.Threshold.Passed {
			for _, violation := range function.Threshold.Violations {
				fmt.Fprintf(stdout, "  %s high complexity for %s: actual=%f limit=%f\n", redLabel("FAILED"), violation.Metric, violation.Actual, violation.Limit)
			}
		}
	}
}

func printFunctionSummary(stdout io.Writer, functions []halstead.FunctionReport) {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Functions")
	fmt.Fprintln(stdout, "--------------------------------")
	fmt.Fprintf(stdout, "Found %d functions.\n", len(functions))

	flagged := 0
	for _, function := range functions {
		if function.Threshold != nil && !function.Threshold.Passed {
			flagged++
		}
	}
	if flagged == 0 {
		fmt.Fprintln(stdout, "No function exceeded the configured thresholds.")
		return
	}

	fmt.Fprintf(stdout, "%s %d function(s) exceeded the configured thresholds.\n", redLabel("FAILED"), flagged)
}

func printJSON(stdout io.Writer, report halstead.AnalysisReport) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
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
	report, err := halstead.AnalyzeASTSource(absPath, []byte(data))
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

func printCounts(stdout io.Writer, values map[string]int) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintln(stdout, "Key:", key, "Value:", values[key])
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

func printThresholdSummary(stdout io.Writer, report halstead.AnalysisReport) {
	if report.Threshold == nil {
		return
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Thresholds")
	fmt.Fprintln(stdout, "--------------------------------")
	if report.Threshold.Passed {
		fmt.Fprintf(stdout, "%s file metrics are within the configured limits.\n", greenLabel("OK"))
	} else {
		fmt.Fprintf(stdout, "%s the file exceeds the configured complexity thresholds.\n", redLabel("FAILED"))
		for _, violation := range report.Threshold.Violations {
			fmt.Fprintf(stdout, "  %s high complexity for %s: actual=%f limit=%f\n", redLabel("ALERT"), violation.Metric, violation.Actual, violation.Limit)
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

func printComparisonSummary(stdout io.Writer, report halstead.AnalysisReport, verbose bool) {
	if report.Comparison == nil {
		return
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Comparison")
	fmt.Fprintln(stdout, "--------------------------------")
	fmt.Fprintf(stdout, "Baseline report: %s\n", report.Comparison.BaselinePath)
	fmt.Fprintf(stdout, "File delta: volume=%f difficulty=%f effort=%f\n", report.Comparison.File.Delta.VolumeDelta, report.Comparison.File.Delta.DifficultyDelta, report.Comparison.File.Delta.EffortDelta)
	if report.Comparison.File.Threshold != nil && !report.Comparison.File.Threshold.Passed {
		for _, violation := range report.Comparison.File.Threshold.Violations {
			fmt.Fprintf(stdout, "  %s file complexity increase for %s: delta=%f limit=%f\n", redLabel("FAILED"), violation.Metric, violation.Delta, violation.Limit)
		}
	}
	if !verbose {
		printComparisonFunctionSummary(stdout, report.Comparison.Functions)
		return
	}
	for _, function := range report.Comparison.Functions {
		fmt.Fprintf(stdout, "%s [%s] delta: volume=%f difficulty=%f effort=%f\n", function.Name, function.Kind, function.Delta.VolumeDelta, function.Delta.DifficultyDelta, function.Delta.EffortDelta)
		if !function.FoundInBase {
			fmt.Fprintln(stdout, "  new function in current report")
		}
		if function.Threshold != nil && !function.Threshold.Passed {
			for _, violation := range function.Threshold.Violations {
				fmt.Fprintf(stdout, "  %s function complexity increase for %s: delta=%f limit=%f\n", redLabel("FAILED"), violation.Metric, violation.Delta, violation.Limit)
			}
		}
	}
}

func printComparisonFunctionSummary(stdout io.Writer, functions []halstead.FunctionComparison) {
	if len(functions) == 0 {
		fmt.Fprintln(stdout, "No changed functions in comparison.")
		return
	}

	changed := 0
	flagged := 0
	for _, function := range functions {
		if function.IsChanged || !function.FoundInBase {
			changed++
		}
		if function.Threshold != nil && !function.Threshold.Passed {
			flagged++
		}
	}

	fmt.Fprintf(stdout, "Compared %d function(s); %d changed or new.\n", len(functions), changed)
	if flagged == 0 {
		fmt.Fprintln(stdout, "No function delta exceeded the configured thresholds.")
		return
	}
	fmt.Fprintf(stdout, "%s %d function delta(s) exceeded the configured thresholds.\n", redLabel("FAILED"), flagged)
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

func redLabel(text string) string {
	return ansiBold + ansiRed + text + ansiReset
}

func greenLabel(text string) string {
	return ansiBold + ansiGreen + text + ansiReset
}
