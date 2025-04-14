package api

type TestPackage struct {
	PackageIdentifier string
	VersionMask       string
}

func NewTestPackage(packageIdentifier string, versionMask string) *TestPackage {
	return &TestPackage{packageIdentifier, versionMask}
}
