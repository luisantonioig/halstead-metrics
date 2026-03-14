package main

import "fmt"

func describe(v any) string {
	switch value := v.(type) {
	case int:
		return fmt.Sprintf("int:%d", value)
	case string:
		return "string:" + value
	default:
		return "unknown"
	}
}

func main() {
	fmt.Println(describe("go"))
}
