package main

import (
	"fmt"

	"./utils"
)

func main() {
	r, err := utils.BuildRing(2, 4, true)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(r)
}
