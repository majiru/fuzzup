package fuzzup

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func errproc(t *testing.T, c chan error) {
	for err := range c {
		t.Fatal(err)
	}
}

func TestBuildURL(t *testing.T) {
	outter := []string{"a", "c", "e", "g"}
	inner := []string{"b", "d", "f"}
	expected := "abcdefg"
	result := buildURL(outter, inner)
	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

func TestRead(t *testing.T) {
	const fuzzin = "a\tb\nc\td\ne\tf\n"
	target := "https://example.com?{{}}=test&test2={{}}"
	r := bufio.NewScanner(strings.NewReader(fuzzin))
	out := make(chan *Record)
	errc := make(chan error)
	go readproc(strings.Split(target, "{{}}"), r, out, errc)
	go errproc(t, errc)
	i := 0
	for rec := range out {
		switch i {
		case 0:
			const e = "https://example.com?a=test&test2=b"
			if rec.url != e {
				t.Errorf("Expected %s got %s", e, rec.url)
			}
		case 1:
			const e = "https://example.com?c=test&test2=d"
			if rec.url != e {
				t.Errorf("Expected %s got %s", e, rec.url)
			}
		case 2:
			const e = "https://example.com?e=test&test2=f"
			if rec.url != e {
				t.Errorf("Expected %s got %s", e, rec.url)
			}
		}
		i++
	}
}

func TestFetch(t *testing.T) {
	testPages := []string{"Test", "Test2", "uniq", "other"}
	testPageContent := [][]byte{[]byte("Test"), []byte("Test"), []byte("Test2"), []byte("Test3")}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch incoming := strings.Trim(r.URL.String(), "/"); incoming {
		case testPages[0]:
			fmt.Fprint(w, string(testPageContent[0]))
		case testPages[1]:
			fmt.Fprint(w, string(testPageContent[1]))
		case testPages[2]:
			fmt.Fprint(w, string(testPageContent[2]))
		case testPages[3]:
			fmt.Fprint(w, string(testPageContent[3]))
		default:
			t.Errorf("Unknown request %s", incoming)
		}
	}))
	defer ts.Close()
	in := make(chan *Record)
	out := make(chan *Record)
	errc := make(chan error)
	go errproc(t, errc)
	go fetchproc(in, out, errc)
	go func() {
		for _, page := range testPages {
			in <- &Record{url: fmt.Sprintf("%s/%s", ts.URL, page)}
		}
		close(in)
	}()
	i := 0
	for r := range out {
		e := &Record{b: testPageContent[i], url: fmt.Sprintf("%s/%s", ts.URL, testPages[i])}
		if e.url != r.url || bytes.Compare(e.b, r.b) != 0 {
			t.Errorf("Expected %q got %q on round %d", e, r, i)
		}
		i++
	}
}

func TestHash(t *testing.T) {
	sample := []string{"Test", "Test", "Test2", "Test3", "Test4"}
	in := make(chan *Record)
	out := make(chan *Record)
	errc := make(chan error)
	go errproc(t, errc)
	go hashproc("", in, out, errc)
	go func() {
		for _, s := range sample {
			in <- &Record{b: []byte(s)}
		}
		close(in)
	}()
	i := 0
	for r := range out {
		e := &Record{}
		switch i {
		case 1:
			e.count = 2
		default:
			e.count = 1
		}
		if r.count != e.count {
			t.Errorf("Expected %d, got %d", e.count, r.count)
		}
		i++
	}
}
