package commandline

import (
	"fmt"
	"net/url"
	"strings"
)

type UriBuilder struct {
	uri url.URL
}

func (b *UriBuilder) OverrideUri(overrideUri *url.URL) {
	scheme := b.uri.Scheme
	host := b.uri.Host
	path := b.uri.Path

	if overrideUri != nil && overrideUri.Scheme != "" {
		scheme = overrideUri.Scheme
	}
	if overrideUri != nil && overrideUri.Host != "" {
		host = overrideUri.Host
	}
	if overrideUri != nil && overrideUri.Path != "" {
		path = overrideUri.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
	}
	uri, _ := url.Parse(fmt.Sprintf("%s://%s%s", scheme, host, path))
	b.uri = *uri
}

func (b UriBuilder) Uri() url.URL {
	return b.uri
}

func NewUriBuilder(uri url.URL) *UriBuilder {
	return &UriBuilder{uri}
}
