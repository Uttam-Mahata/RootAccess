package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashFlag creates a SHA-256 hash of the flag
func HashFlag(flag string) string {
	hash := sha256.Sum256([]byte(flag))
	return hex.EncodeToString(hash[:])
}

// VerifyFlag compares a submitted flag against a stored hash
func VerifyFlag(submittedFlag, storedHash string) bool {
	submittedHash := HashFlag(submittedFlag)
	return submittedHash == storedHash
}
