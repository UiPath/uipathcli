package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ExternalAuthenticator struct {
	Config ExternalAuthenticatorConfig
}

func (a ExternalAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	input, err := json.Marshal(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error serializing input authenticator context for external authenticator '%s': %v", a.Config.Name, err))
	}

	path, err := a.getAuthenticatorPath()
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error invoking external authenticator '%s' using path '%s': %v", a.Config.Name, a.Config.Path, err))
	}
	cmd := exec.Command(path)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewReader(input)

	err = cmd.Run()
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error invoking external authenticator '%s' using path '%s', output: %s: %v", a.Config.Name, a.Config.Path, stderr.String(), err))
	}

	var result AuthenticatorResult
	err = json.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error parsing output authenticator context: %v", err))
	}
	return result
}

func (a ExternalAuthenticator) getAuthenticatorPath() (string, error) {
	if filepath.IsAbs(a.Config.Path) {
		return a.Config.Path, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Error retrieving executable path: %v", err)
	}
	directory := filepath.Dir(executable)
	return filepath.Join(directory, a.Config.Path), nil
}
