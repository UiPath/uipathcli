package auth

type JwtInfo struct {
	PartId string
}

func NewJwtInfo(partId string) *JwtInfo {
	return &JwtInfo{partId}
}
