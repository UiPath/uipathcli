package config

// PluginConfigStore is an abstraction for reading plugin configuration
type PluginConfigStore interface {
	Read() ([]byte, error)
}
