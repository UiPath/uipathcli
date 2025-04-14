package studio

import "encoding/xml"

type junitReport struct {
	XMLName    xml.Name               `xml:"testsuites"`
	TestSuites []junitReportTestSuite `xml:"testsuite"`
}

type junitReportTestSuite struct {
	XMLName   xml.Name              `xml:"testsuite"`
	Id        string                `xml:"id,attr"`
	Name      string                `xml:"name,attr"`
	Time      float64               `xml:"time,attr"`
	Package   string                `xml:"package,attr"`
	Tests     int                   `xml:"tests,attr"`
	Failures  int                   `xml:"failures,attr"`
	Errors    int                   `xml:"errors,attr"`
	Cancelled int                   `xml:"cancelled,attr"`
	SystemOut string                `xml:"system-out"`
	TestCases []junitReportTestCase `xml:"testcase"`
}

type junitReportTestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	Name      string   `xml:"name,attr"`
	Time      float64  `xml:"time,attr"`
	Status    string   `xml:"status,attr"`
	Classname string   `xml:"classname,attr"`
	SystemOut string   `xml:"system-out"`
}
