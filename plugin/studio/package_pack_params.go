package studio

type packagePackParams struct {
	Source         string
	Destination    string
	PackageVersion string
	AutoVersion    bool
	OutputType     string
	SplitOutput    bool
	ReleaseNotes   string
}

func newPackagePackParams(
	source string,
	destination string,
	packageVersion string,
	autoVersion bool,
	outputType string,
	splitOutput bool,
	releaseNotes string) *packagePackParams {
	return &packagePackParams{source, destination, packageVersion, autoVersion, outputType, splitOutput, releaseNotes}
}
