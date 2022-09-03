package executor

type TokenProvider interface {
	GetToken(request TokenRequest) (string, error)
}
