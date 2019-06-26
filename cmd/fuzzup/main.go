package main

import (
	"fmt"
	"log"
	"os"

	"github.com/majiru/fuzzup"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s target [wordlist]\n", os.Args[0])
	os.Exit(1)
}

func main() {
	switch n := len(os.Args); {
	case n < 2, n > 3:
		usage()
	case n == 2:
		log.Fatal(fuzzup.Fuzz(os.Args[1], os.Stdin))
	case n == 3:
		f, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.Fatal(fuzzup.Fuzz(os.Args[1], f))
	}
}
