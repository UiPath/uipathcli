package studio

type nuspec struct {
	Title   string
	Version string
}

func newNuspec(title string, version string) *nuspec {
	return &nuspec{title, version}
}
