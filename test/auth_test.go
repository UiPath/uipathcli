package test

import (
	"strings"
	"testing"
)

func TestNoAuth(t *testing.T) {
	config := `
profiles:
  - name: default
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	header := result.RequestHeader["authorization"]
	if header != "" {
		t.Errorf("Expected no Authorization header, but got: %v", header)
	}
}

func TestBearerAuthIdentityErrorResponse(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      clientId: failure-client-id
      clientSecret: failure-client-secret
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(500, "Internal Server Error").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if result.Error.Error() != "Error retrieving bearer token: Service returned status code '500' and body 'Internal Server Error'" {
		t.Errorf("Expected error from identity, but got: %v", result.Error)
	}
}

func TestBearerAuthSuccess(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      clientId: success-client-id
      clientSecret: success-client-secret
      properties:
        custom: myvalue
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	authorization := result.RequestHeader["authorization"]
	if authorization != "Bearer my-jwt-access-token" {
		t.Errorf("Expected bearer token from identity, but got: %v", authorization)
	}
}

func TestBearerAuthForwardsRequestId(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      clientId: success-client-id
      clientSecret: success-client-secret
      properties:
        custom: myvalue
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	requestId := result.RequestHeader["x-request-id"]
	if requestId == "" {
		t.Errorf("Expected x-request-id header, but got: %v", requestId)
	}
}

func TestBearerAuthWithInvalidIdentityUriConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      clientId: success-client-id
      clientSecret: success-client-secret
      uri: -invalid-uri%
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if !strings.Contains(result.Error.Error(), "Error parsing identity uri") {
		t.Errorf("Expected identity uri parsing error, but got: %v", result.Error)
	}
}

func TestBearerAuthWithInvalidIdentityUriParameter(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      clientId: success-client-id
      clientSecret: success-client-secret
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--identity-uri", ":invalid"}, context)

	if !strings.Contains(result.Error.Error(), "Error parsing identity-uri argument") {
		t.Errorf("Expected identity uri parsing error, but got: %v", result.Error)
	}
}

func TestBearerAuthTokenIsCached(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      clientId: cached-client-id
      clientSecret: cached-client-secret
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()
	RunCli([]string{"myservice", "ping"}, context)

	context2 := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(500, "Internal Server Error").
		Build()
	result := RunCli([]string{"myservice", "ping"}, context2)

	err := result.Error
	if err != nil {
		t.Errorf("Expected no call to identity, but call failed: %v", err)
	}
	authorization := result.RequestHeader["authorization"]
	if authorization != "Bearer my-jwt-access-token" {
		t.Errorf("Expected bearer token from identity, but got: %v", authorization)
	}
}

func TestBearerAuthTokenRetrievedWhenExpired(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      clientId: cached-client-id-expired-token
      clientSecret: cached-client-secret-expired-token
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 10, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()
	RunCli([]string{"myservice", "ping"}, context)

	context2 := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(500, "Internal Server Error").
		Build()
	result := RunCli([]string{"myservice", "ping"}, context2)

	if result.Error.Error() != "Error retrieving bearer token: Service returned status code '500' and body 'Internal Server Error'" {
		t.Errorf("Expected identity call, but got: %v", result.Error)
	}
}

func TestAuthPATSuccessfully(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      pat: rt_mypat
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	expected := "Bearer rt_mypat"
	header := result.RequestHeader["authorization"]
	if header != expected {
		t.Errorf("Expected PAT in Authorization header, got: %v", header)
	}
}

func TestCacheClearShowsConfirmation(t *testing.T) {
	context := NewContextBuilder().
		Build()
	result := RunCli([]string{"config", "cache", "clear"}, context)

	if result.StdOut != "Cache has been successfully cleared\n" {
		t.Errorf("Expected cache cleared message, but got: %v", result.StdOut)
	}
}

func TestCacheClearRemovesCachedAuthToken(t *testing.T) {
	config := `
profiles:
  - name: default
    auth:
      clientId: cleared-client-id
      clientSecret: cleared-client-secret
`
	definition := `
paths:
  /ping:
    get:
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()
	RunCli([]string{"myservice", "ping"}, context)

	context2 := NewContextBuilder().
		Build()
	RunCli([]string{"config", "cache", "clear"}, context2)

	context3 := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(500, "Internal Server Error").
		Build()
	result := RunCli([]string{"myservice", "ping"}, context3)

	if result.Error == nil {
		t.Error("Expected call to identity but did not receive")
	}
}
