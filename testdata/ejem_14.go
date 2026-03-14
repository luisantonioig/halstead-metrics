package main

import "fmt"

type person struct {
	name string
	age  int
}

func main() {
	p := person{name: "Ana", age: 30}
	fmt.Println(p.name, p.age)
}
