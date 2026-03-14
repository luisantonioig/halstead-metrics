package halstead

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExamplesWorkWithBothAnalyzers(t *testing.T) {
	for i := 1; i <= 20; i++ {
		path := filepath.Join("testdata", exampleName(i))
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}

		astMetrics, err := AnalyzeAST(src)
		if err != nil {
			t.Fatalf("AnalyzeAST(%s): %v", path, err)
		}
		if astMetrics.TotalOperators == 0 || astMetrics.TotalOperands == 0 {
			t.Fatalf("AnalyzeAST(%s) returned empty metrics: %+v", path, astMetrics)
		}
	}
}

func exampleName(i int) string {
	if i < 10 {
		return "ejem_0" + string(rune('0'+i)) + ".go"
	}
	if i == 10 {
		return "ejem_10.go"
	}
	return "ejem_" + string(rune('0'+(i/10))) + string(rune('0'+(i%10))) + ".go"
}
