package converter

import (
	"net/url"
	"testing"
)

func TestRemovesTrailingSlash(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com/"), "/my-service")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-service" {
		t.Errorf("Did not remove trailing slash, got: %v", uri)
	}
}

func TestAddsMissingSlashSeparator(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "my-service")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-service" {
		t.Errorf("Did not add missing slash separator, got: %v", uri)
	}
}

func TestFormatPathReplacesPlaceholder(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/{organization}/{tenant}/my-service")

	builder.FormatPath("organization", "my-org")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-org/{tenant}/my-service" {
		t.Errorf("Did not replace placeholder, got: %v", uri)
	}
}

func TestFormatPathReplacesPlaceholderInBaseUri(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com/{organization}/{tenant}/"), "/my-service")

	builder.FormatPath("organization", "my-org")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-org/{tenant}/my-service" {
		t.Errorf("Did not replace placeholder, got: %v", uri)
	}
}

func TestFormatPathReplacesPlaceholderWithEscapedPathValue(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com/{organization}/{tenant}/"), "/my-service")

	builder.FormatPath("organization", "my org")
	builder.FormatPath("tenant", "{my%tenant}")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my%20org/%7Bmy%25tenant%7D/my-service" {
		t.Errorf("Did not replace placeholder, got: %v", uri)
	}
}

func TestFormatPathReplacesPlaceholderWithEscapedQueryStringKey(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/my-service")

	builder.AddQueryString("folder&", "Shared")
	builder.AddQueryString("id?", 10)

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-service?folder%26=Shared&id%3F=10" {
		t.Errorf("Did not replace placeholder, got: %v", uri)
	}
}

func TestFormatPathReplacesPlaceholderWithEscapedQueryStringValue(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/my-service")

	builder.AddQueryString("folder", "&Shared?")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-service?folder=%26Shared%3F" {
		t.Errorf("Did not replace placeholder, got: %v", uri)
	}
}

func TestFormatPathReplacesMultiplePlaceholders(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com/{organization}/{tenant}"), "/my-service")

	builder.FormatPath("organization", "my-org")
	builder.FormatPath("tenant", "my-tenant")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-org/my-tenant/my-service" {
		t.Errorf("Did not replace placeholder, got: %v", uri)
	}
}

func TestFormatPathDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) { FormatPathDataTypes(t, "my-value", "my-value") })
	t.Run("Integer", func(t *testing.T) { FormatPathDataTypes(t, 1, "1") })
	t.Run("Number", func(t *testing.T) { FormatPathDataTypes(t, 1.4, "1.4") })
	t.Run("Boolean", func(t *testing.T) { FormatPathDataTypes(t, true, "true") })
	t.Run("StringArray", func(t *testing.T) {
		FormatPathDataTypes(t, []string{"my-value-1", "my-value-2"}, "my-value-1,my-value-2")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		FormatPathDataTypes(t, []int{1, 2, 3}, "1,2,3")
	})
	t.Run("NumberArray", func(t *testing.T) {
		FormatPathDataTypes(t, []float64{1.3, 2.2, 3.5}, "1.3,2.2,3.5")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		FormatPathDataTypes(t, []bool{true, false, true}, "true,false,true")
	})
}
func FormatPathDataTypes(t *testing.T, value interface{}, expected string) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/{param}")

	builder.FormatPath("param", value)

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/"+expected {
		t.Errorf("Did not format data type properly, got: %v", uri)
	}
}

func TestFormatPathEscapeStringValue(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/{param}")

	builder.FormatPath("param", "my#value")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my%23value" {
		t.Errorf("Did not escape path properly, got: %v", uri)
	}
}

func TestFormatPathPreservesSlash(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/{param}")

	builder.FormatPath("param", "my/value")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my/value" {
		t.Errorf("Did not preserve slash in path, got: %v", uri)
	}
}

func TestFormatPathEscapeStringArrayValue(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/{param}")

	builder.FormatPath("param", []string{"my/value-1", "my-value-2"})

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my%2Fvalue-1,my-value-2" {
		t.Errorf("Did not escape path properly, got: %v", uri)
	}
}

func TestAddQueryString(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/my-service")

	builder.AddQueryString("filter", "my-value")

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-service?filter=my-value" {
		t.Errorf("Did not add querystring, got: %v", uri)
	}
}

func TestAddMultipleQueryStringParameters(t *testing.T) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/my-service")

	builder.AddQueryString("skip", 1)
	builder.AddQueryString("take", 5)

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-service?skip=1&take=5" {
		t.Errorf("Did not add querystring, got: %v", uri)
	}
}

func TestQueryStringDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) { QueryStringDataTypes(t, "my-value", "?param=my-value") })
	t.Run("Integer", func(t *testing.T) { QueryStringDataTypes(t, 1000000000, "?param=1000000000") })
	t.Run("Number", func(t *testing.T) { QueryStringDataTypes(t, 99.341231923, "?param=99.341231923") })
	t.Run("Boolean", func(t *testing.T) { QueryStringDataTypes(t, false, "?param=false") })
	t.Run("StringArray", func(t *testing.T) {
		QueryStringDataTypes(t, []string{"my-value-1", "my-value-2"}, "?param=my-value-1&param=my-value-2")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		QueryStringDataTypes(t, []int{100, 200, 300}, "?param=100&param=200&param=300")
	})
	t.Run("NumberArray", func(t *testing.T) {
		QueryStringDataTypes(t, []float64{0.001, 0.002}, "?param=0.001&param=0.002")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		QueryStringDataTypes(t, []bool{true, false}, "?param=true&param=false")
	})
}
func QueryStringDataTypes(t *testing.T, value interface{}, expected string) {
	builder := NewUriBuilder(toUrl("https://cloud.uipath.com"), "/my-service")

	builder.AddQueryString("param", value)

	uri := builder.Build()
	if uri != "https://cloud.uipath.com/my-service"+expected {
		t.Errorf("Did not format data type for query string properly, got: %v", uri)
	}
}

func toUrl(uri string) url.URL {
	result, _ := url.Parse(uri)
	return *result
}
