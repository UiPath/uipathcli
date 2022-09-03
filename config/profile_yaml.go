package config

type profileYaml struct {
	Name         string            `yaml:"name"`
	Uri          urlYaml           `yaml:"uri"`
	Path         map[string]string `yaml:"path"`
	Query        map[string]string `yaml:"query"`
	Header       map[string]string `yaml:"header"`
	ClientId     string            `yaml:"clientId"`
	ClientSecret string            `yaml:"clientSecret"`
	Insecure     bool              `yaml:"insecure"`
	Debug        bool              `yaml:"debug"`
}
