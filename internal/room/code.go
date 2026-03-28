// Package room handles room lifecycle, peer state, and code generation.
package room

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// charset is the alphabet used for room codes.
// Excludes visually ambiguous characters: 0, O, 1, I, L.
const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// CodeLength is the fixed length of all Tsuna room codes.
const CodeLength = 6

// GenerateCode produces a cryptographically random 6-character room code
// using the safe charset above. Examples: "SAKURA", "MX4T9Q", "BRN7WA".
func GenerateCode() (string, error) {
	n := big.NewInt(int64(len(charset)))
	var b strings.Builder
	b.Grow(CodeLength)

	for i := 0; i < CodeLength; i++ {
		idx, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", err
		}
		b.WriteByte(charset[idx.Int64()])
	}

	return b.String(), nil
}

// ValidateCode returns true if the code is exactly CodeLength characters
// and consists entirely of characters from the charset.
func ValidateCode(code string) bool {
	if len(code) != CodeLength {
		return false
	}
	for _, c := range strings.ToUpper(code) {
		if !strings.ContainsRune(charset, c) {
			return false
		}
	}
	return true
}

// NormalizeCode upper-cases and trims whitespace from a user-supplied code.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
