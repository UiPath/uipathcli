package api

type TestSet struct {
	Id        int
	TestCases []TestCase
	Packages  []TestPackage
}

func NewTestSet(id int, testCases []TestCase, testPackages []TestPackage) *TestSet {
	return &TestSet{id, testCases, testPackages}
}
