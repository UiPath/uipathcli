package unpack

type solutionUnpackParams struct {
	Source      string
	Destination string
}

func newSolutionUnpackParams(source string, destination string) *solutionUnpackParams {
	return &solutionUnpackParams{source, destination}
}
