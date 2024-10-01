package commandline

import "fmt"

// The FlagType is an enum of the supported flag definition types.
type FlagType int

const (
	FlagTypeString FlagType = iota + 1
	FlagTypeInteger
	FlagTypeBoolean
	FlagTypeStringArray
)

func (t FlagType) String() string {
	switch t {
	case FlagTypeString:
		return "string"
	case FlagTypeInteger:
		return "integer"
	case FlagTypeBoolean:
		return "boolean"
	case FlagTypeStringArray:
		return "stringArray"
	}
	panic(fmt.Sprintf("Unknown flag type: %d", int(t)))
}
