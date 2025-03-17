package studio

type nuspec struct {
	Id      string
	Title   string
	Version string
}

func newNuspec(id string, title string, version string) *nuspec {
	return &nuspec{id, title, version}
}
