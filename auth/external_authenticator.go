package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// The ExternalAuthenticator invokes a configurable executable which is providing the
// authentication credentials.
//
// The ExternalAuthenticator serializes the AuthenticatorContext and passes it on standard input
// to the external executable. The executable performs the authentication and returns
// the AuthenticatorResult on standard output.
//
// Example: Authenticator which uses kubernetes to retrieve clientId, clientSecret
// https://github.com/UiPath/uipathcli-authenticator-k8s
type ExternalAuthenticator struct {
	config ExternalAuthenticatorConfig
}

func (a ExternalAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	input, err := json.Marshal(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error serializing input authenticator context for external authenticator '%s': %w", a.config.Name, err))
	}

	path, err := a.getAuthenticatorPath()
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error invoking external authenticator '%s' using path '%s': %w", a.config.Name, a.config.Path, err))
	}
	cmd := exec.Command(path)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewReader(input)

	err = cmd.Run()
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error invoking external authenticator '%s' using path '%s', output: %s: %w", a.config.Name, a.config.Path, stderr.String(), err))
	}

	var result AuthenticatorResult
	err = json.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error parsing output authenticator context: %w", err))
	}
	return result
}

func (a ExternalAuthenticator) getAuthenticatorPath() (string, error) {
	if filepath.IsAbs(a.config.Path) {
		return a.config.Path, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Error retrieving executable path: %w", err)
	}
	directory := filepath.Dir(executable)
	return filepath.Join(directory, a.config.Path), nil
}

func NewExternalAuthenticator(config ExternalAuthenticatorConfig) *ExternalAuthenticator {
	return &ExternalAuthenticator{}
}
