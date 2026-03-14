package main

import "fmt"

func classify(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	default:
		return "C"
	}
}

func main() {
	fmt.Println(classify(85))
}
