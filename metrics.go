package halstead

import "math"

const deltaEpsilon = 1e-9

// Metrics represents the raw token counts and derived Halstead metrics for one analysis run.
type Metrics struct {
	Name           string
	Operators      map[string]int
	Operands       map[string]int
	TotalOperators int
	TotalOperands  int
}

// MetricsSummary is a JSON-friendly view of Metrics including derived values.
type MetricsSummary struct {
	Name                    string         `json:"name"`
	Operators               map[string]int `json:"operators"`
	Operands                map[string]int `json:"operands"`
	TotalOperators          int            `json:"total_operators"`
	TotalOperands           int            `json:"total_operands"`
	DifferentOperators      int            `json:"different_operators"`
	DifferentOperands       int            `json:"different_operands"`
	ProgramVocabulary       int            `json:"program_vocabulary"`
	ProgramLength           int            `json:"program_length"`
	CalculatedProgramLength float64        `json:"calculated_program_length"`
	Volume                  float64        `json:"volume"`
	Difficulty              float64        `json:"difficulty"`
	Effort                  float64        `json:"effort"`
	TimeRequiredToProgram   float64        `json:"time_required_to_program"`
	NumberOfDeliveredBugs   float64        `json:"number_of_delivered_bugs"`
}

// Position identifies a 1-based source location.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// FunctionReport contains function-level metrics and source location data.
type FunctionReport struct {
	Name      string            `json:"name"`
	Kind      string            `json:"kind"`
	Start     Position          `json:"start"`
	End       Position          `json:"end"`
	Metrics   MetricsSummary    `json:"metrics"`
	Threshold *ThresholdOutcome `json:"threshold,omitempty"`
}

// AnalysisReport is a machine-readable result for one analyzed source file.
type AnalysisReport struct {
	Analyzer   string             `json:"analyzer"`
	Path       string             `json:"path,omitempty"`
	File       MetricsSummary     `json:"file"`
	Functions  []FunctionReport   `json:"functions"`
	Threshold  *ThresholdOutcome  `json:"threshold,omitempty"`
	Comparison *ComparisonOutcome `json:"comparison,omitempty"`
}

// Thresholds configures pass/fail checks for reports and functions.
type Thresholds struct {
	MaxVolume     float64
	MaxDifficulty float64
	MaxEffort     float64
}

// ThresholdViolation describes one exceeded threshold.
type ThresholdViolation struct {
	Metric string  `json:"metric"`
	Actual float64 `json:"actual"`
	Limit  float64 `json:"limit"`
}

// ThresholdOutcome records whether a report passed configured thresholds.
type ThresholdOutcome struct {
	Passed     bool                 `json:"passed"`
	Violations []ThresholdViolation `json:"violations,omitempty"`
}

// DeltaThresholds configures limits for metric growth relative to a baseline report.
type DeltaThresholds struct {
	MaxVolumeDelta     float64
	MaxDifficultyDelta float64
	MaxEffortDelta     float64
}

// ComparisonViolation describes one exceeded delta threshold.
type ComparisonViolation struct {
	Metric   string  `json:"metric"`
	Delta    float64 `json:"delta"`
	Limit    float64 `json:"limit"`
	Baseline float64 `json:"baseline"`
	Current  float64 `json:"current"`
}

// MetricsDelta captures baseline/current changes for key derived metrics.
type MetricsDelta struct {
	VolumeDelta     float64 `json:"volume_delta"`
	DifficultyDelta float64 `json:"difficulty_delta"`
	EffortDelta     float64 `json:"effort_delta"`
}

// FunctionComparison compares one function against its baseline counterpart.
type FunctionComparison struct {
	Name        string                      `json:"name"`
	Kind        string                      `json:"kind"`
	Baseline    *MetricsSummary             `json:"baseline,omitempty"`
	Current     MetricsSummary              `json:"current"`
	Delta       MetricsDelta                `json:"delta"`
	FoundInBase bool                        `json:"found_in_baseline"`
	IsChanged   bool                        `json:"is_changed"`
	Threshold   *ComparisonThresholdOutcome `json:"threshold,omitempty"`
}

// ComparisonThresholdOutcome records whether delta checks passed.
type ComparisonThresholdOutcome struct {
	Passed     bool                  `json:"passed"`
	Violations []ComparisonViolation `json:"violations,omitempty"`
}

// ComparisonOutcome compares a report with a baseline report.
type ComparisonOutcome struct {
	BaselinePath string                      `json:"baseline_path,omitempty"`
	File         FunctionComparison          `json:"file"`
	Functions    []FunctionComparison        `json:"functions"`
	Threshold    *ComparisonThresholdOutcome `json:"threshold,omitempty"`
}

func (m Metrics) DifferentOperators() int {
	return len(m.Operators)
}

func (m Metrics) DifferentOperands() int {
	return len(m.Operands)
}

func (m Metrics) ProgramVocabulary() int {
	return m.DifferentOperators() + m.DifferentOperands()
}

func (m Metrics) ProgramLength() int {
	return m.TotalOperators + m.TotalOperands
}

func (m Metrics) CalculatedProgramLength() float64 {
	return safeLogTerm(m.DifferentOperators()) + safeLogTerm(m.DifferentOperands())
}

func (m Metrics) Volume() float64 {
	vocabulary := m.ProgramVocabulary()
	if vocabulary == 0 {
		return 0
	}
	return float64(m.ProgramLength()) * math.Log2(float64(vocabulary))
}

func (m Metrics) Difficulty() float64 {
	if m.DifferentOperands() == 0 {
		return 0
	}
	return (float64(m.DifferentOperators()) / 2) * (float64(m.TotalOperands) / float64(m.DifferentOperands()))
}

func (m Metrics) Effort() float64 {
	return m.Difficulty() * m.Volume()
}

func (m Metrics) TimeRequiredToProgram() float64 {
	return m.Effort() / 18
}

func (m Metrics) NumberOfDeliveredBugs() float64 {
	effort := m.Effort()
	if effort == 0 {
		return 0
	}
	return math.Pow(effort, 2.0/3.0) / 3000
}

func safeLogTerm(n int) float64 {
	if n == 0 {
		return 0
	}
	return float64(n) * math.Log2(float64(n))
}

// Summary returns a JSON-friendly snapshot with derived values precomputed.
func (m Metrics) Summary() MetricsSummary {
	return MetricsSummary{
		Name:                    m.Name,
		Operators:               cloneCounts(m.Operators),
		Operands:                cloneCounts(m.Operands),
		TotalOperators:          m.TotalOperators,
		TotalOperands:           m.TotalOperands,
		DifferentOperators:      m.DifferentOperators(),
		DifferentOperands:       m.DifferentOperands(),
		ProgramVocabulary:       m.ProgramVocabulary(),
		ProgramLength:           m.ProgramLength(),
		CalculatedProgramLength: m.CalculatedProgramLength(),
		Volume:                  m.Volume(),
		Difficulty:              m.Difficulty(),
		Effort:                  m.Effort(),
		TimeRequiredToProgram:   m.TimeRequiredToProgram(),
		NumberOfDeliveredBugs:   m.NumberOfDeliveredBugs(),
	}
}

func cloneCounts(src map[string]int) map[string]int {
	dst := make(map[string]int, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

// Evaluate reports whether the summary passed the configured thresholds.
func (m MetricsSummary) Evaluate(thresholds Thresholds) *ThresholdOutcome {
	violations := collectViolations(m, thresholds)
	if !thresholds.Enabled() && len(violations) == 0 {
		return nil
	}
	return &ThresholdOutcome{
		Passed:     len(violations) == 0,
		Violations: violations,
	}
}

// Enabled reports whether any threshold is configured.
func (t Thresholds) Enabled() bool {
	return t.MaxVolume > 0 || t.MaxDifficulty > 0 || t.MaxEffort > 0
}

func collectViolations(summary MetricsSummary, thresholds Thresholds) []ThresholdViolation {
	var violations []ThresholdViolation
	if thresholds.MaxVolume > 0 && summary.Volume > thresholds.MaxVolume {
		violations = append(violations, ThresholdViolation{
			Metric: "volume",
			Actual: summary.Volume,
			Limit:  thresholds.MaxVolume,
		})
	}
	if thresholds.MaxDifficulty > 0 && summary.Difficulty > thresholds.MaxDifficulty {
		violations = append(violations, ThresholdViolation{
			Metric: "difficulty",
			Actual: summary.Difficulty,
			Limit:  thresholds.MaxDifficulty,
		})
	}
	if thresholds.MaxEffort > 0 && summary.Effort > thresholds.MaxEffort {
		violations = append(violations, ThresholdViolation{
			Metric: "effort",
			Actual: summary.Effort,
			Limit:  thresholds.MaxEffort,
		})
	}
	return violations
}

// Enabled reports whether any delta threshold is configured.
func (t DeltaThresholds) Enabled() bool {
	return t.MaxVolumeDelta > 0 || t.MaxDifficultyDelta > 0 || t.MaxEffortDelta > 0
}

// CompareReports computes deltas between a baseline report and a current report.
func CompareReports(baseline, current AnalysisReport) ComparisonOutcome {
	comparison := ComparisonOutcome{
		BaselinePath: baseline.Path,
		File: newFunctionComparison(
			"<file>",
			"file",
			&baseline.File,
			current.File,
			true,
		),
	}

	baselineFunctions := make(map[string]FunctionReport, len(baseline.Functions))
	for _, function := range baseline.Functions {
		baselineFunctions[functionComparisonKey(function.Name, function.Kind)] = function
	}

	for _, function := range current.Functions {
		key := functionComparisonKey(function.Name, function.Kind)
		baseFunction, ok := baselineFunctions[key]
		if ok {
			comparison.Functions = append(comparison.Functions, newFunctionComparison(
				function.Name,
				function.Kind,
				&baseFunction.Metrics,
				function.Metrics,
				true,
			))
			continue
		}
		comparison.Functions = append(comparison.Functions, newFunctionComparison(
			function.Name,
			function.Kind,
			nil,
			function.Metrics,
			false,
		))
	}

	return comparison
}

// Evaluate applies delta thresholds to a comparison outcome.
func (c *ComparisonOutcome) Evaluate(thresholds DeltaThresholds) {
	if !thresholds.Enabled() {
		return
	}
	fileOutcome := c.File.evaluate(thresholds)
	c.File.Threshold = fileOutcome

	var failed bool
	if fileOutcome != nil && !fileOutcome.Passed {
		failed = true
	}
	for i := range c.Functions {
		outcome := c.Functions[i].evaluate(thresholds)
		c.Functions[i].Threshold = outcome
		if outcome != nil && !outcome.Passed {
			failed = true
		}
	}
	c.Threshold = &ComparisonThresholdOutcome{Passed: !failed}
}

func (f FunctionComparison) evaluate(thresholds DeltaThresholds) *ComparisonThresholdOutcome {
	violations := collectComparisonViolations(f, thresholds)
	if !thresholds.Enabled() && len(violations) == 0 {
		return nil
	}
	return &ComparisonThresholdOutcome{
		Passed:     len(violations) == 0,
		Violations: violations,
	}
}

func collectComparisonViolations(function FunctionComparison, thresholds DeltaThresholds) []ComparisonViolation {
	var violations []ComparisonViolation
	if thresholds.MaxVolumeDelta > 0 && function.Delta.VolumeDelta > thresholds.MaxVolumeDelta {
		violations = append(violations, newComparisonViolation("volume", function.Delta.VolumeDelta, thresholds.MaxVolumeDelta, function))
	}
	if thresholds.MaxDifficultyDelta > 0 && function.Delta.DifficultyDelta > thresholds.MaxDifficultyDelta {
		violations = append(violations, newComparisonViolation("difficulty", function.Delta.DifficultyDelta, thresholds.MaxDifficultyDelta, function))
	}
	if thresholds.MaxEffortDelta > 0 && function.Delta.EffortDelta > thresholds.MaxEffortDelta {
		violations = append(violations, newComparisonViolation("effort", function.Delta.EffortDelta, thresholds.MaxEffortDelta, function))
	}
	return violations
}

func newComparisonViolation(metric string, delta, limit float64, function FunctionComparison) ComparisonViolation {
	var baseline, current float64
	switch metric {
	case "volume":
		current = function.Current.Volume
		if function.Baseline != nil {
			baseline = function.Baseline.Volume
		}
	case "difficulty":
		current = function.Current.Difficulty
		if function.Baseline != nil {
			baseline = function.Baseline.Difficulty
		}
	case "effort":
		current = function.Current.Effort
		if function.Baseline != nil {
			baseline = function.Baseline.Effort
		}
	}
	return ComparisonViolation{
		Metric:   metric,
		Delta:    delta,
		Limit:    limit,
		Baseline: baseline,
		Current:  current,
	}
}

func newFunctionComparison(name, kind string, baseline *MetricsSummary, current MetricsSummary, found bool) FunctionComparison {
	delta := MetricsDelta{
		VolumeDelta:     current.Volume - summaryValue(baseline, "volume"),
		DifficultyDelta: current.Difficulty - summaryValue(baseline, "difficulty"),
		EffortDelta:     current.Effort - summaryValue(baseline, "effort"),
	}
	return FunctionComparison{
		Name:        name,
		Kind:        kind,
		Baseline:    baseline,
		Current:     current,
		FoundInBase: found,
		IsChanged:   !found || metricsDeltaChanged(delta),
		Delta:       delta,
	}
}

func summaryValue(summary *MetricsSummary, metric string) float64 {
	if summary == nil {
		return 0
	}
	switch metric {
	case "volume":
		return summary.Volume
	case "difficulty":
		return summary.Difficulty
	case "effort":
		return summary.Effort
	default:
		return 0
	}
}

func functionComparisonKey(name, kind string) string {
	return kind + ":" + name
}

// ChangedOnly returns a copy that includes only changed or newly added functions.
func (c ComparisonOutcome) ChangedOnly() ComparisonOutcome {
	filtered := c
	filtered.Functions = nil
	for _, function := range c.Functions {
		if function.IsChanged {
			filtered.Functions = append(filtered.Functions, function)
		}
	}
	return filtered
}

func metricsDeltaChanged(delta MetricsDelta) bool {
	return abs(delta.VolumeDelta) > deltaEpsilon ||
		abs(delta.DifficultyDelta) > deltaEpsilon ||
		abs(delta.EffortDelta) > deltaEpsilon
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
