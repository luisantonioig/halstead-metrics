package main

import (
	"fmt"
	"os"
	"sort"

	halstead "github.com/luisantonioig/halstead-metrics"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: halstead <archivo.go>")
		os.Exit(2)
	}

	path := os.Args[1]
	src, err := os.ReadFile(path)
	check(err)

	metrics, err := halstead.AnalyzeAST(src)
	check(err)
	printMetrics(metrics)
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
