package config

import (
	"strings"
	"testing"
)

func TestCryptoEncryptDecrypt(t *testing.T) {
	orig := "sk-ant-api03-abcdef123456789"
	enc, err := EncryptSecret(orig)
	if err != nil {
		t.Fatalf("EncryptSecret failed: %v", err)
	}
	if !strings.HasPrefix(enc, "enc:v1:") {
		t.Fatalf("expected enc:v1: prefix, got %s", enc)
	}
	if enc == orig {
		t.Fatalf("ciphertext must not match plaintext")
	}

	dec, err := DecryptSecret(enc)
	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}
	if dec != orig {
		t.Fatalf("expected decrypted %q, got %q", orig, dec)
	}
}

func TestCryptoPlaintextFallback(t *testing.T) {
	raw := "sk-raw-unencrypted-key"
	dec, err := DecryptSecret(raw)
	if err != nil {
		t.Fatalf("expected no error on raw plaintext, got %v", err)
	}
	if dec != raw {
		t.Fatalf("expected %q, got %q", raw, dec)
	}
}

func TestCryptoEmpty(t *testing.T) {
	enc, err := EncryptSecret("")
	if err != nil || enc != "" {
		t.Fatalf("expected empty result for empty secret, got %q, err %v", enc, err)
	}
	dec, err := DecryptSecret("")
	if err != nil || dec != "" {
		t.Fatalf("expected empty result for empty decrypted, got %q, err %v", dec, err)
	}
}
