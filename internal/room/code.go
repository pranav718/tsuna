package room

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

const CodeLength = 6

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

func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
