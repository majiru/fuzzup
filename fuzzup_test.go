package fuzzup

import (
	"fmt"
	"testing"
)

func TestParseTarget(t *testing.T) {
	url := []string{"https://example.com/vuln?", "=", "&extra=val"}
	parts := ParseTarget(fmt.Sprintf("%s{{}}%s{{}}%s", url[0], url[1], url[2]))
	if len(parts) == 0 {
		t.Errorf("Regexp failed to parse url")
	}
	if len(parts) != len(url) {
		t.Errorf("Number of input strings does not match output.\nIn %q, out %q", url, parts)
	}
	for i := range url {
		if url[i] != parts[i] {
			t.Errorf("Expected %s, got %s", url[i], parts[i])
		}
	}
}
