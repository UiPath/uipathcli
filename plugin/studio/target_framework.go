package studio

import (
	"errors"
	"runtime"
)

type TargetFramework int

const (
	TargetFrameworkCrossPlatform TargetFramework = iota + 1
	TargetFrameworkWindows
	TargetFrameworkLegacy
)

func (t TargetFramework) IsWindowsOnly() bool {
	return t == TargetFrameworkWindows || t == TargetFrameworkLegacy
}

func (t TargetFramework) IsSupported() (bool, error) {
	if t.IsWindowsOnly() && runtime.GOOS != "windows" {
		return false, errors.New("UiPath Studio Projects which target windows-only are not support on linux devices. Build the project on windows or change the target framework to cross-platform.")
	}
	return true, nil
}
