package main

import "fmt"

func main() {
	ch := make(chan string, 1)
	ch <- "ready"

	select {
	case msg := <-ch:
		fmt.Println(msg)
	default:
		fmt.Println("empty")
	}
}
