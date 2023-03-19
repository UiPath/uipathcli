package commandline

import (
	"github.com/urfave/cli/v2"
)

type flagBuilder struct {
	flags map[string]cli.Flag
	order []string
}

func (b *flagBuilder) AddFlag(flag cli.Flag) {
	name := flag.Names()[0]
	if _, found := b.flags[name]; !found {
		b.flags[name] = flag
		b.order = append(b.order, name)
	}
}

func (b *flagBuilder) AddFlags(flags []cli.Flag) {
	for _, flag := range flags {
		b.AddFlag(flag)
	}
}

func (b flagBuilder) ToList() []cli.Flag {
	flags := []cli.Flag{}
	for _, name := range b.order {
		flags = append(flags, b.flags[name])
	}
	return flags
}

func newFlagBuilder() *flagBuilder {
	return &flagBuilder{map[string]cli.Flag{}, []string{}}
}
