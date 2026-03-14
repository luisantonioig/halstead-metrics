package halstead

import "math"

// Metrics represents the raw token counts and derived Halstead metrics for one analysis run.
type Metrics struct {
	Name           string
	Operators      map[string]int
	Operands       map[string]int
	TotalOperators int
	TotalOperands  int
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
