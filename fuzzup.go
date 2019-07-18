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
	"sync"
)

type Record struct {
	url   string
	b     []byte
	count int
}

//map[string]*Record
var ledger sync.Map = sync.Map{}

func buildURL(outter []string, inner []string) string {
	var s string
	for i := range inner {
		s = s + outter[i] + inner[i]
	}
	return s + outter[len(outter)-1]
}

func readproc(target []string, scanner *bufio.Scanner, c chan *Record, errc chan error) {
	for i := 1; scanner.Scan(); i++ {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) != len(target)-1 {
			errc <- fmt.Errorf("Line %d: Expected %d fields, line has %d", i, len(target)-1, len(parts))
			continue
		}
		c <- &Record{url: buildURL(target, parts)}
	}
	close(c)
}

func fetchproc(in chan *Record, out chan *Record, errc chan error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	for rec := range in {
		r, err := client.Get(rec.url)
		if err != nil {
			errc <- err
			continue
		}
		rec.b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			errc <- err
			continue
		}
		r.Body.Close()
		out <- rec
	}
	close(out)
}

func hashproc(filter string, in chan *Record, out chan *Record, errc chan error) {
	rxp := regexp.MustCompile(filter)
	for rec := range in {
		rec.b = rxp.ReplaceAll(rec.b, []byte(""))
		hashStr := fmt.Sprintf("%x", sha256.Sum256(rec.b))
		val, loaded := ledger.LoadOrStore(hashStr, rec)
		rec, ok := val.(*Record)
		if !ok {
			errc <- fmt.Errorf("fetch: cast to *Record failed")
		}
		if loaded {
			rec.count++
		} else {
			rec.count = 1
		}
		out <- rec
	}
	close(errc)
	close(out)
}

func Fuzz(target string, filter string, wl io.Reader) error {
	readout := make(chan *Record)
	fetchout := make(chan *Record)
	hashout := make(chan *Record)
	errchan := make(chan error)
	go readproc(strings.Split(target, "{{}}"), bufio.NewScanner(wl), readout, errchan)
	go fetchproc(readout, fetchout, errchan)
	go hashproc(filter, fetchout, hashout, errchan)

	for {
		select {
		case err := <-errchan:
			return err
		case rec := <-hashout:
			if rec.count == 1 {
				fmt.Printf("Url %s generated a unique page hash\n", rec.url)
			}
		}
	}

	return nil
}
