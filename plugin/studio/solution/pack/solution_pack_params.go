package pack

type solutionPackParams struct {
	Source       string
	Destination  string
	SolutionId   string
	SolutionName string
}

func newSolutionPackParams(source string, destination string, solutionId string, solutionName string) *solutionPackParams {
	return &solutionPackParams{source, destination, solutionId, solutionName}
}
