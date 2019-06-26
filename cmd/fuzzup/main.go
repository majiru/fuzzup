package main

import (
	"os"
	"fmt"
	"log"

	"github.com/majiru/fuzzup"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s target [wordlist]\n", os.Args[0])
	os.Exit(1)
}

func main() {
	var fuzzer *fuzzup.Fuzzer
	switch n := len(os.Args); {
	case n < 2, n > 3:
		usage()
	case n == 2:
		fuzzer = fuzzup.NewFuzzer(os.Args[1], os.Stdin)
	case n == 3:
		f, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		fuzzer = fuzzup.NewFuzzer(os.Args[1], f)
	}
	log.Fatal(fuzzer.Fuzz())
}