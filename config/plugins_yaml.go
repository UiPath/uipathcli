package config

type pluginsYaml struct {
	Authenticators []authenticatorYaml `yaml:"authenticators"`
}
