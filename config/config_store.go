package config

// ConfigStore provides an abstraction for reading and writing the config file.
type ConfigStore interface {
	Read() ([]byte, error)
	Write(data []byte) error
}
