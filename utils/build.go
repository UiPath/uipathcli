// Package utils contains command functionality which is reused across
// multiple other packages.
package utils

// Version variable is overridden during build time by providing the linker flag:
// -ldflags="-X github.com/UiPath/uipathcli/utils.Version=1.2.3"
var Version = "main"
