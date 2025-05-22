package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type JwtParser struct {
}

func (p JwtParser) Parse(token string) (*JwtInfo, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("Invalid token")
	}
	payload, err := p.base64UrlDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("Could not decode token payload: %w", err)
	}

	var payloadJson jwtPayloadJson
	if err := json.Unmarshal(payload, &payloadJson); err != nil {
		return nil, fmt.Errorf("Could not deserialize token payload: %w", err)
	}
	return NewJwtInfo(payloadJson.PrtId), nil
}

func (p JwtParser) base64UrlDecode(str string) ([]byte, error) {
	if missingPadding := len(str) % 4; missingPadding != 0 {
		str += strings.Repeat("=", 4-missingPadding)
	}
	return base64.URLEncoding.DecodeString(str)
}

func NewJwtParser() *JwtParser {
	return &JwtParser{}
}

type jwtPayloadJson struct {
	PrtId string `json:"prt_id"`
}
