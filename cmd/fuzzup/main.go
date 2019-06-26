package main

import (
	"fmt"
	"log"
	"os"

	"github.com/majiru/fuzzup"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s target regexp [wordlist]\n", os.Args[0])
	os.Exit(1)
}

func main() {
	switch n := len(os.Args); {
	case n < 3, n > 4:
		usage()
	case n == 3:
		log.Fatal(fuzzup.Fuzz(os.Args[1], os.Args[2], os.Stdin))
	case n == 4:
		f, err := os.Open(os.Args[3])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.Fatal(fuzzup.Fuzz(os.Args[1], os.Args[2], f))
	}
}
