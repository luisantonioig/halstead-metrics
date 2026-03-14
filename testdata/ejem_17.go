package main

import "fmt"

func cleanup(label string) {
	fmt.Println("cleanup:", label)
}

func main() {
	defer cleanup("demo")
	fmt.Println("run")
}
