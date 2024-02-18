package main

import "fmt"

func main() {
	for _, i := range []int{1, 2, 3} {
		fmt.Println(i, "is even", i%2)
	}
}
