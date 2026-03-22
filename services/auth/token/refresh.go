package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateRefreshToken returns a cryptographically random opaque token (hex-encoded, 32 bytes = 64 chars).
func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	hash = HashToken(raw)
	return raw, hash, nil
}

// HashToken returns the SHA-256 hex digest of a token.
// Used for storage and lookup - no need for bcrypt here since the token itself has high entropy (256 bits).
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
