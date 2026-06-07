package auth

import (
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	var out struct {
		Email string `json:"email"`
	}
	err := decodeJSON(strings.NewReader(`{"email":"a@b.com"}`), &out)
	if err != nil || out.Email != "a@b.com" {
		t.Fatalf("err=%v out=%#v", err, out)
	}
	err = decodeJSON(strings.NewReader(`{`), &out)
	if err == nil {
		t.Fatal("expected decode error")
	}
}