package config

type profileYaml struct {
	Name     string                 `yaml:"name"`
	Uri      urlYaml                `yaml:"uri"`
	Path     map[string]string      `yaml:"path"`
	Query    map[string]string      `yaml:"query"`
	Header   map[string]string      `yaml:"header"`
	Auth     map[string]interface{} `yaml:"auth"`
	Insecure bool                   `yaml:"insecure"`
	Debug    bool                   `yaml:"debug"`
}
