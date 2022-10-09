package config

import "net/url"

type Config struct {
	Uri      *url.URL
	Path     map[string]string
	Query    map[string]string
	Header   map[string]string
	Auth     AuthConfig
	Insecure bool
	Debug    bool
}

type AuthConfig struct {
	Type   string
	Config map[string]interface{}
}
