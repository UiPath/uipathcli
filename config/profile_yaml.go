package config

type profileYaml struct {
	Name     string                 `yaml:"name"`
	Uri      urlYaml                `yaml:"uri,omitempty"`
	Path     map[string]string      `yaml:"path,omitempty"`
	Query    map[string]string      `yaml:"query,omitempty"`
	Header   map[string]string      `yaml:"header,omitempty"`
	Auth     map[string]interface{} `yaml:"auth,omitempty"`
	Insecure bool                   `yaml:"insecure,omitempty"`
	Debug    bool                   `yaml:"debug,omitempty"`
}
