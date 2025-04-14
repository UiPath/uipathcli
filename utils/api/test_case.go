package api

type TestCase struct {
	Id                int
	Name              string
	PackageIdentifier string
}

func NewTestCase(id int, name string, packageIdentifier string) *TestCase {
	return &TestCase{id, name, packageIdentifier}
}
