package fuzzup

import (
	"fmt"
	"io"
	"bufio"
	"regexp"
	"net/http"
	"strings"
	"crypto/sha256"
)

type Fuzzer struct {
	target []string
	wl *bufio.Scanner
}

func ParseTarget(s string) (out []string) {
	re := regexp.MustCompile(`{{}}`)
	matches := re.FindAllStringIndex(s, -1)
	if matches == nil {
		return
	}
	out = append(out, s[:matches[0][0]])
	for i := 1; i < len(matches); i++ {
		out = append(out, s[matches[i-1][1]:matches[i][0]])
	}
	out = append(out, s[matches[len(matches)-1][1]:])
	return
}

func NewFuzzer(target string, wl io.Reader) *Fuzzer {
	return &Fuzzer{ParseTarget(target), bufio.NewScanner(wl)}
}

func buildURL(outter []string, inner []string) string {
	var s string
	for i := range inner {
		s = fmt.Sprintf("%s%s%s", s, outter[i], inner[i])
	}
	return fmt.Sprintf("%s%s", s, outter[len(outter)-1])
}

func (f *Fuzzer) read(c chan string, errc chan error) {
	for i := 1; f.wl.Scan(); i++ {
		line := f.wl.Text()
		parts := strings.Split(line, "\t")
		if len(parts) != len(f.target)-1 {
			errc <- fmt.Errorf("Line %d: Expected %d fields, line has %d", i, len(f.target)-1, len(parts))
			continue
		}
		c <- buildURL(f.target, parts)
	}
	close(c)
}

type Record struct {
	url string
	count int
}

func (f *Fuzzer) fetch(in chan string, out chan Record, errc chan error) {
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

func (f *Fuzzer) Fuzz() error {
	urlchan := make(chan string)
	recchan := make(chan Record)
	errchan := make(chan error)
	go f.read(urlchan, errchan)
	go f.fetch(urlchan, recchan, errchan)

	for rec := range recchan {
		select {
		case err := <- errchan:
			return err
		default:
			if rec.count == 1 {
				fmt.Printf("Url %s generated a unique page hash\n", rec.url)
			}
		}
	}

	return nil
}