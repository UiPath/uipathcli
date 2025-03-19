package api

type Release struct {
	Id   int
	Name string
}

func NewRelease(id int, name string) *Release {
	return &Release{id, name}
}
