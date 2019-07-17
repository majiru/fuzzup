package fuzzup

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildURL(t *testing.T) {
	outter := []string{"a", "c", "e", "g"}
	inner := []string{"b", "d", "f"}

	expected := fmt.Sprintf("%s%s%s%s%s%s%s", outter[0], inner[0], outter[1], inner[1], outter[2], inner[2], outter[3])
	result := buildURL(outter, inner)
	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

func TestRead(t *testing.T) {
	const fuzzin = "a\tb\nc\td\ne\tf\n"
	target := "https://example.com?{{}}=test&test2={{}}"
	r := bufio.NewScanner(strings.NewReader(fuzzin))
	out := make(chan string)
	errc := make(chan error)
	go func() {
		read(strings.Split(target, "{{}}"), r, out, errc)
		close(errc)
	}()
	go func() {
		for err := range errc {
			t.Errorf("Got error %s when reading", err.Error())
		}
	}()
	i := 0
	for url := range out {
		switch i {
		case 0:
			const e = "https://example.com?a=test&test2=b"
			if url != e {
				t.Errorf("Expected %s got %s", e, url)
			}
		case 1:
			const e = "https://example.com?c=test&test2=d"
			if url != e {
				t.Errorf("Expected %s got %s", e, url)
			}
		case 2:
			const e = "https://example.com?e=test&test2=f"
			if url != e {
				t.Errorf("Expected %s got %s", e, url)
			}
		}
		i++
	}
}

func TestFetch(t *testing.T) {
	testPages := []string{"Test", "Test2", "uniq", "other"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch incoming := strings.Trim(r.URL.String(), "/"); incoming {
		case testPages[0]:
			fmt.Fprintln(w, "Test")
		case testPages[1]:
			fmt.Fprintln(w, "Test")
		case testPages[2]:
			fmt.Fprintln(w, "Test2")
		case testPages[3]:
			fmt.Fprintln(w, "Test3")
		default:
			t.Errorf("Unknown request %s", incoming)
		}
	}))
	defer ts.Close()
	in := make(chan string)
	out := make(chan Record)
	errc := make(chan error)
	go func() {
		for err := range errc {
			t.Errorf("%q", err)
		}
	}()
	go fetch("", in, out, errc)
	go func() {
		for _, page := range testPages {
			in <- fmt.Sprintf("%s/%s", ts.URL, page)
		}
		close(in)
	}()
	i := 0
	for r := range out {
		eurl := fmt.Sprintf("%s/%s", ts.URL, testPages[i])
		e := Record{url: eurl}
		switch i {
		case 0, 2, 3:
			e.count = 1
		case 1:
			e.url = fmt.Sprintf("%s/%s", ts.URL, testPages[0])
			e.count = 2
		}
		if e != r {
			t.Errorf("Expected %q got %q", e, r)
		}
		i++
	}
}
