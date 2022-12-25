package cache

type Cache interface {
	Get(key string) (string, float32)
	Set(key string, value string, expiresIn float32)
}
