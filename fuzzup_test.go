package fuzzup

import (
	"fmt"
	"testing"
)


func TestBuildURL(t *testing.T) {
	outter := []string{"a", "c", "e", "g"}
	inner := []string{"b", "d", "f"}

	expected := fmt.Sprintf("%s%s%s%s%s%s%s", outter[0], inner[0], outter[1], inner[1], outter[2], inner[2], outter[3])
	result := buildURL(outter, inner)
	if result !=  expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}