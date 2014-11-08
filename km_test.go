package main

import "testing"

func TestParseConfig(t *testing.T) {
	if _, err := parseConfig("/afilethatprobablynotexists"); err == nil {
		t.Error("expected error on invalid filename")
	}
}
