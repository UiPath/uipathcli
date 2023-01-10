package commandline

import (
	"net/url"
	"testing"
)

func TestUriReturnsInitialUri(t *testing.T) {
	uri := toUri("http://localhost:1234/test")
	builder := NewUriBuilder(uri)

	result := builder.Uri()
	if result != uri {
		t.Errorf("Should return initial uri, but got: %v", result)
	}
}

func TestOverrideUri(t *testing.T) {
	t.Run("TestOverrideUriSetsSchema", func(t *testing.T) {
		OverrideUri(t, "http://localhost:5000/mypath", "https://", "https://localhost:5000/mypath")
	})
	t.Run("TestOverrideUriSetsHostname", func(t *testing.T) {
		OverrideUri(t, "http://localhost:5000/mypath", "http://cloud.uipath.com", "http://cloud.uipath.com/mypath")
	})
	t.Run("TestOverrideUriSetsSchemaAndHostname", func(t *testing.T) {
		OverrideUri(t, "http://localhost:5000/mypath", "https://cloud.uipath.com", "https://cloud.uipath.com/mypath")
	})
	t.Run("TestOverrideUriSetsPath", func(t *testing.T) {
		OverrideUri(t, "http://localhost:5000/mypath", "/otherpath", "http://localhost:5000/otherpath")
	})
	t.Run("TestOverrideUriTrimsSlash", func(t *testing.T) {
		OverrideUri(t, "http://localhost:5000/mypath/", "/otherpath/", "http://localhost:5000/otherpath")
	})
	t.Run("TestOverrideUriMultiplePathSegments", func(t *testing.T) {
		OverrideUri(t, "https://cloud.uipath.com/mypath/myresource/", "otherpath", "https://cloud.uipath.com/otherpath")
	})
	t.Run("TestOverrideAddsPath", func(t *testing.T) {
		OverrideUri(t, "https://cloud.uipath.com", "otherpath", "https://cloud.uipath.com/otherpath")
	})
	t.Run("TestOverridePort", func(t *testing.T) {
		OverrideUri(t, "https://cloud.uipath.com/mypath", "http://localhost:1234", "http://localhost:1234/mypath")
	})
}

func OverrideUri(t *testing.T, initial string, override string, expected string) {
	uri := toUri(initial)
	builder := NewUriBuilder(uri)

	overrideUri := toUri(override)
	builder.OverrideUri(&overrideUri)

	result := builder.Uri()
	if result != toUri(expected) {
		t.Errorf("Should return expected uri %v, but got: %v", expected, result)
	}
}

func TestOverrideMultipleUris(t *testing.T) {
	uri := toUri("https://cloud.uipath.com/mypath/")
	builder := NewUriBuilder(uri)

	overrideUri := toUri("https://cloud.uipath.com/first")
	builder.OverrideUri(&overrideUri)
	overrideUri2 := toUri("https://alpha.uipath.com")
	builder.OverrideUri(&overrideUri2)

	result := builder.Uri()
	if result != toUri("https://alpha.uipath.com/first") {
		t.Errorf("Should return expected uri, but got: %v", result)
	}
}

func toUri(rawURL string) url.URL {
	uri, _ := url.Parse(rawURL)
	return *uri
}
