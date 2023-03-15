package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

type secretGenerator struct{}

func (g secretGenerator) GeneratePkce() (string, string) {
	random := []byte(g.randomString(32))
	codeVerifier := g.base64Encode(random)

	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := g.base64Encode(hash[:])
	return codeVerifier, codeChallenge
}

func (g secretGenerator) GenerateState() string {
	random := []byte(g.randomString(32))
	return g.base64Encode(random)
}

func (g secretGenerator) base64Encode(value []byte) string {
	result := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(value)
	result = strings.ReplaceAll(result, "+", "-")
	result = strings.ReplaceAll(result, "/", "_")
	return result
}

func (g secretGenerator) randomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Errorf("Could not get cryptographically secure random numbers: %w", err))
	}
	return fmt.Sprintf("%x", b)[:length]
}

func newSecretGenerator() *secretGenerator {
	return &secretGenerator{}
}
