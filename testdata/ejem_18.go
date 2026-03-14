package main

import "fmt"

func main() {
	values := make(chan int, 1)
	go func() {
		values <- 7
	}()
	fmt.Println(<-values)
}
