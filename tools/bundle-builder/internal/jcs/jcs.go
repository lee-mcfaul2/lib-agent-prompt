package jcs

import (
	"crypto/sha256"
	"encoding/hex"

	canonicaljson "github.com/gibson042/canonicaljson-go"
)

// Marshal serializes v to RFC 8785 JCS canonical JSON.
func Marshal(v any) ([]byte, error) {
	return canonicaljson.Marshal(v)
}

// Sha256OfJCS returns the SHA-256 digest of the canonical JCS encoding of v as "sha256:<hex>".
func Sha256OfJCS(v any) (string, error) {
	b, err := Marshal(v)
	if err != nil {
		return "", err
	}
	return sha256Hex(b), nil
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(h[:])
}
