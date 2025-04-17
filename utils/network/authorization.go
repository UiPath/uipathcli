package network

// Authorization represents the values of the http authorization header.
//
// Example:
// Authorization.Type: "Bearer"
// Authorization.Value: "<jwt-bearer-token>"
type Authorization struct {
	Type  string
	Value string
}

func NewAuthorization(type_ string, value string) *Authorization {
	return &Authorization{type_, value}
}
