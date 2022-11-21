package provider

import (
	"testing"
)

func TestProvider(t *testing.T) {
	if err := NewProvider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
