package api

type TestCaseAssertion struct {
	Message   string
	Succeeded bool
}

func NewTestCaseAssertion(message string, succeeded bool) *TestCaseAssertion {
	return &TestCaseAssertion{message, succeeded}
}
