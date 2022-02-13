package modules

import (
	"strings"
	"testing"
)

func TestHttpCheck(t *testing.T) {

	h := &HTTPModule{}

	args := make(map[string]interface{})

	// Missing 'url'
	err := h.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing url")
	}
	if !strings.Contains(err.Error(), "missing 'url'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Valid target
	args["url"] = "https://github.com"
	err = h.Check(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
}
