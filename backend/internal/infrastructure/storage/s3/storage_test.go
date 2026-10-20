package s3

import (
	"encoding/hex"
	"testing"
)

// Vector from the AWS SigV4 documentation ("Derive a signing key").
func TestSigningKey(t *testing.T) {
	got := hex.EncodeToString(signingKey("wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "20150830", "us-east-1", "iam"))
	want := "c4afb1cc5771d871763a393e44b703571b55cc28424d1a5e86da6ed3c154a4b9"
	if got != want {
		t.Errorf("signingKey = %s, want %s", got, want)
	}
}

func TestEscapeKey(t *testing.T) {
	if got := escapeKey("public/a b/c+d.png"); got != "public/a%20b/c+d.png" {
		t.Errorf("escapeKey = %q", got)
	}
}
