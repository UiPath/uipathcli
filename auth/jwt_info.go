package auth

type JwtInfo struct {
	PrtId string
}

func NewJwtInfo(prtId string) *JwtInfo {
	return &JwtInfo{prtId}
}
