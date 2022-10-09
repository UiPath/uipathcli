package plugins

type pluginsYaml struct {
	Authenticators []authenticatorYaml `yaml:"authenticators"`
}
