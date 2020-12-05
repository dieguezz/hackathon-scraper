package main

import "fmt"

func pollas() (int, int, int, int) {
	return 1, 3, 4, 5
}

func main() {
	a, b, c, d := pollas()
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(d)
}
