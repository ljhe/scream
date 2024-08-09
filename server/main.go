package main

import (
	"fmt"
)

func main() {
	Test(1, 2, func(c, d int) {
		fmt.Println("c:", c)
		fmt.Println("d:", d)
	})
}

func Test(a, b int, f func(c, d int)) {
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	f(a+1, b+1)
}
