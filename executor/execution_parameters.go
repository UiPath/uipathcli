package executor

import "github.com/UiPath/uipathcli/parser"

type ExecutionParameters []ExecutionParameter

func (p ExecutionParameters) Path() []ExecutionParameter {
	return p.filter(parser.ParameterInPath)
}

func (p ExecutionParameters) Query() []ExecutionParameter {
	return p.filter(parser.ParameterInQuery)
}

func (p ExecutionParameters) Header() []ExecutionParameter {
	return p.filter(parser.ParameterInHeader)
}

func (p ExecutionParameters) Body() []ExecutionParameter {
	return p.filter(parser.ParameterInBody)
}

func (p ExecutionParameters) Form() []ExecutionParameter {
	return p.filter(parser.ParameterInForm)
}

func (p ExecutionParameters) filter(in string) []ExecutionParameter {
	result := []ExecutionParameter{}
	for _, p := range p {
		if p.In == in {
			result = append(result, p)
		}
	}
	return result
}
