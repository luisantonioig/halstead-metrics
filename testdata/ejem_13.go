package main

import "fmt"

func frequencies(words []string) map[string]int {
	counts := map[string]int{}
	for _, word := range words {
		counts[word]++
	}
	return counts
}

func main() {
	fmt.Println(frequencies([]string{"go", "go", "ast"}))
}
