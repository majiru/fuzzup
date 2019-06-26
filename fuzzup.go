package fuzzup

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
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

func fetch(filter string, in chan string, out chan Record, errc chan error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
    client := &http.Client{Transport: tr}
	ledger := make(map[string]Record)
	rxp := regexp.MustCompile(filter)
	for url := range in {
		r, err := client.Get(url)
		if err != nil {
			errc <- err
			continue
		}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errc <- err
			continue
		}
		r.Body.Close()
		b = rxp.ReplaceAll(b, []byte(""))
		hashStr := fmt.Sprintf("%x", sha256.Sum256(b))
		if rec, ok := ledger[hashStr]; ok {
			rec.count++
			out <- rec
		} else {
			rec = Record{url, 1}
			ledger[hashStr] = rec
			out <- rec
		}
	}
	close(errc)
	close(out)
}

func Fuzz(target string, filter string, wl io.Reader) error {
	urlchan := make(chan string)
	recchan := make(chan Record)
	errchan := make(chan error)
	go read(strings.Split(target, "{{}}"), bufio.NewScanner(wl), urlchan, errchan)
	go fetch(filter, urlchan, recchan, errchan)

	for {
		select {
		case err := <-errchan:
			return err
		case rec := <-recchan:
			if rec.count == 1 {
				fmt.Printf("Url %s generated a unique page hash\n", rec.url)
			}
		}
	}

	return nil
}
