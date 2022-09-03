package parser

type Parser interface {
	Parse(name string, data []byte) (*Definition, error)
}
