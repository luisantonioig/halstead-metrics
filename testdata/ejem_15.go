package main

import "fmt"

type counter int

func (c counter) next() counter {
	return c + 1
}

func main() {
	var start counter = 2
	fmt.Println(start.next())
}
