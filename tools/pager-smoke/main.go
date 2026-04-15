package main

import (
	"flag"
	"fmt"
)

func main() {
	count := flag.Int("n", 1000, "number of lines to print")
	prefix := flag.String("prefix", "line", "line prefix")
	flag.Parse()

	for index := 1; index <= *count; index++ {
		fmt.Printf("%s %d\n", *prefix, index)
	}
}
