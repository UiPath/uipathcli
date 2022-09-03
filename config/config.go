package config

import "net/url"

type Config struct {
	Uri          *url.URL
	Path         map[string]string
	Query        map[string]string
	Header       map[string]string
	ClientId     string
	ClientSecret string
	Insecure     bool
	Debug        bool
}
