package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) > 1 {
		code, err := strconv.Atoi(os.Args[1])
		if err == nil {
			fmt.Printf("exit code: %d\n", code)
			os.Exit(code)
		}
	}
	os.Exit(0)
}
