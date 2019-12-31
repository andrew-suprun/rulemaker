package main

import "fmt"

func foo() (result []int) {
	result = append(result, 1)
	result = append(result, 2)
	result = append(result, 3)
	return result
}

func main() {
	fmt.Printf("%v\n", foo())
}
