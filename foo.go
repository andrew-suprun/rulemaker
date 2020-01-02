package main

import "fmt"

func foo() (result []int) {
	result = append(result, 1)
	result = append(result, 2)
	result = append(result, 3)
	return result
}

func mainX() {
	fmt.Printf("%v\n", foo())
}
