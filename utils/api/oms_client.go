package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils/converter"
	"github.com/UiPath/uipathcli/utils/network"
)

type OmsClient struct {
	baseUri  url.URL
	token    *auth.AuthToken
	settings network.HttpClientSettings
	logger   log.Logger
}

func (c OmsClient) GetOrganizationInfo(organizationId string) (*Organization, error) {
	uri := converter.NewUriBuilder(c.baseUri, "/{organizationId}/organization_/api/organization/{organizationId}/AllInfo").
		FormatPath("organizationId", organizationId).
		Build()
	header := http.Header{
		"Content-Type": {"application/json"},
	}
	request := network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)

	client := network.NewHttpClient(c.logger, c.settings)
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Organization Management Service returned status code '%d' and body '%s'", response.StatusCode, string(responseBody))
	}

	var responseJson organizationInfoJson
	err = json.Unmarshal(responseBody, &responseJson)
	if err != nil {
		return nil, fmt.Errorf("Error parsing json response: %w", err)
	}
	return c.convertToOrganization(responseJson), nil
}

func (c OmsClient) convertToOrganization(response organizationInfoJson) *Organization {
	tenants := []Tenant{}
	for _, tenant := range response.Tenants {
		tenants = append(tenants, *NewTenant(tenant.Id, tenant.Name))
	}
	return NewOrganization(response.Organization.Id, response.Organization.Name, tenants)
}

func (c OmsClient) toAuthorization(token *auth.AuthToken) *network.Authorization {
	if token == nil {
		return nil
	}
	return network.NewAuthorization(token.Type, token.Value)
}

func NewOmsClient(
	baseUri url.URL,
	token *auth.AuthToken,
	settings network.HttpClientSettings,
	logger log.Logger,
) *OmsClient {
	return &OmsClient{
		baseUri,
		token,
		settings,
		logger,
	}
}

type organizationInfoJson struct {
	Organization organizationJson
	Tenants      []tenantJson
}

type organizationJson struct {
	Id   string `json:"id"`
	Name string `json:"logicalName"`
}

type tenantJson struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
