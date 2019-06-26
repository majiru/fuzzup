package fuzzup

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func buildURL(outter []string, inner []string) string {
	var s string
	for i := range inner {
		s = fmt.Sprintf("%s%s%s", s, outter[i], inner[i])
	}
	return fmt.Sprintf("%s%s", s, outter[len(outter)-1])
}

func read(target []string, scanner *bufio.Scanner, c chan string, errc chan error) {
	for i := 1; scanner.Scan(); i++ {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) != len(target)-1 {
			errc <- fmt.Errorf("Line %d: Expected %d fields, line has %d", i, len(target)-1, len(parts))
			continue
		}
		c <- buildURL(target, parts)
	}
	close(c)
}

type Record struct {
	url   string
	count int
}

func fetch(in chan string, out chan Record, errc chan error) {
	ledger := make(map[string]Record)
	h := sha256.New()
	for url := range in {
		h.Reset()
		r, err := http.Get(url)
		if err != nil {
			errc <- err
			continue
		}
		_, err = io.Copy(h, r.Body)
		if err != nil {
			errc <- err
			continue
		}
		hashStr := fmt.Sprintf("%x", h.Sum(nil))
		if rec, ok := ledger[hashStr]; ok {
			rec.count++
			out <- rec
		} else {
			rec = Record{url, 1}
			ledger[hashStr] = rec
			out <- rec
		}
		r.Body.Close()
	}
	close(out)
}

func Fuzz(target string, wl io.Reader) error {
	urlchan := make(chan string)
	recchan := make(chan Record)
	errchan := make(chan error)
	go read(strings.Split(target, "{{}}"), bufio.NewScanner(wl), urlchan, errchan)
	go fetch(urlchan, recchan, errchan)

	for rec := range recchan {
		select {
		case err := <-errchan:
			return err
		default:
			if rec.count == 1 {
				fmt.Printf("Url %s generated a unique page hash\n", rec.url)
			}
		}
	}

	return nil
}
