package main

import (
	"fmt"
	"math/rand"
	"os"
)

func main() {
	N := rand.Intn(100)
	fmt.Println("N =", N)

	if N < 50 {
		fmt.Println("< 50. PASS")
	} else {
		fmt.Println(">= 50. FAIL")
		os.Exit(N)
	}
}
