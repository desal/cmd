package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 4 {
		return
	}

	fmt.Fprintf(os.Stdout, os.Args[1])
	fmt.Fprintf(os.Stderr, os.Args[2])
	exitCode, _ := strconv.ParseInt(os.Args[3], 10, 32)
	os.Exit(int(exitCode))
}
