package fuzzup

import (
	"bytes"
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

func (f *Fuzzer) Fuzz() error {
	client := http.Client{}
	var sum []byte
	for i := 1; f.wl.Scan(); i++ {
		h := sha256.New()
		line := f.wl.Text()
		parts := strings.Split(line, "\t")
		if len(parts) != len(f.target)-1 {
			return fmt.Errorf("Line %d: Expected %d fields, line has %d", i, len(f.target)-1, len(parts))
		}
		url := buildURL(f.target, parts)
		r, err := client.Get(url)
		if err != nil {
			return err
		}
		_, err = io.Copy(h, r.Body)
		if err != nil {
			return err
		}
		r.Body.Close()
		newsum := h.Sum(nil)
		if bytes.Compare(sum, newsum) != 0 && i != 1{
			fmt.Printf("URL %s seems like it might be interesting", url)
		}
		sum = newsum
	}
	return nil
}