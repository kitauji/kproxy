package kproxy

import (
	"testing"
)

func TestPraseHost(t *testing.T) {
	tests := map[string]string{
		"example.com":          "example.com",
		"example.com:8080":     "example.com",
		"www.example.com:8080": "www.example.com",
	}

	for host, expect := range tests {
		actual := parseHostname(host)
		if actual != expect {
			t.Errorf("Hostname should be %s, but %s", actual, expect)
		}
	}
}
