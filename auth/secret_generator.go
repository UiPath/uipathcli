package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type SecretGenerator struct{}

func (g SecretGenerator) GeneratePkce() (string, string) {
	random := []byte(g.randomString(32))
	codeVerifier := g.base64Encode(random)

	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := g.base64Encode(hash[:])
	return codeVerifier, codeChallenge
}

func (g SecretGenerator) GenerateState() string {
	random := []byte(g.randomString(32))
	return g.base64Encode(random)
}

func (g SecretGenerator) base64Encode(value []byte) string {
	result := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(value)
	result = strings.ReplaceAll(result, "+", "-")
	result = strings.ReplaceAll(result, "/", "_")
	return result
}

func (g SecretGenerator) randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}
