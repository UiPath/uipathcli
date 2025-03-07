package studio

import "encoding/xml"

type nuspecXml struct {
	XMLName  xml.Name                 `xml:"package"`
	Metadata nuspecPackageMetadataXml `xml:"metadata"`
}

type nuspecPackageMetadataXml struct {
	Title   string `xml:"title"`
	Version string `xml:"version"`
}
