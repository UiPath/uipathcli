package studio

type Nuspec struct {
	Id      string
	Title   string
	Version string
}

func NewNuspec(id string, title string, version string) *Nuspec {
	return &Nuspec{id, title, version}
}
