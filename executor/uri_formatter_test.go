package executor

import (
	"net/url"
	"testing"
)

func TestRemovesTrailingSlash(t *testing.T) {
	formatter := newUriFormatter(toUrl("https://cloud.uipath.com/"), "/my-service")

	uri := formatter.Uri()
	if uri != "https://cloud.uipath.com/my-service" {
		t.Errorf("Did not remove trailing slash, got: %v", uri)
	}
}

func TestAddsMissingSlashSeparator(t *testing.T) {
	formatter := newUriFormatter(toUrl("https://cloud.uipath.com"), "my-service")

	uri := formatter.Uri()
	if uri != "https://cloud.uipath.com/my-service" {
		t.Errorf("Did not add missing slash separator, got: %v", uri)
	}
}

func TestFormatPathReplacesPlaceholder(t *testing.T) {
	formatter := newUriFormatter(toUrl("https://cloud.uipath.com"), "/{organization}/{tenant}/my-service")

	formatter.FormatPath(*NewExecutionParameter("organization", "my-org", "path"))

	uri := formatter.Uri()
	if uri != "https://cloud.uipath.com/my-org/{tenant}/my-service" {
		t.Errorf("Did not replace placeholder, got: %v", uri)
	}
}

func TestFormatPathReplacesMultiplePlaceholders(t *testing.T) {
	formatter := newUriFormatter(toUrl("https://cloud.uipath.com"), "/{organization}/{tenant}/my-service")

	formatter.FormatPath(*NewExecutionParameter("organization", "my-org", "path"))
	formatter.FormatPath(*NewExecutionParameter("tenant", "my-tenant", "path"))

	uri := formatter.Uri()
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
	formatter := newUriFormatter(toUrl("https://cloud.uipath.com"), "/{param}")

	formatter.FormatPath(*NewExecutionParameter("param", value, "path"))

	uri := formatter.Uri()
	if uri != "https://cloud.uipath.com/"+expected {
		t.Errorf("Did not format data type properly, got: %v", uri)
	}
}

func TestAddQueryString(t *testing.T) {
	formatter := newUriFormatter(toUrl("https://cloud.uipath.com"), "/my-service")

	parameters := []ExecutionParameter{
		*NewExecutionParameter("filter", "my-value", "query"),
	}
	formatter.AddQueryString(parameters)

	uri := formatter.Uri()
	if uri != "https://cloud.uipath.com/my-service?filter=my-value" {
		t.Errorf("Did not add querystring, got: %v", uri)
	}
}

func TestAddMultipleQueryStringParameters(t *testing.T) {
	formatter := newUriFormatter(toUrl("https://cloud.uipath.com"), "/my-service")

	parameters := []ExecutionParameter{
		*NewExecutionParameter("skip", 1, "query"),
		*NewExecutionParameter("take", 5, "query"),
	}
	formatter.AddQueryString(parameters)

	uri := formatter.Uri()
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
	formatter := newUriFormatter(toUrl("https://cloud.uipath.com"), "/my-service")

	parameters := []ExecutionParameter{
		*NewExecutionParameter("param", value, "query"),
	}
	formatter.AddQueryString(parameters)

	uri := formatter.Uri()
	if uri != "https://cloud.uipath.com/my-service"+expected {
		t.Errorf("Did not format data type for query string properly, got: %v", uri)
	}
}

func toUrl(uri string) url.URL {
	result, _ := url.Parse(uri)
	return *result
}
