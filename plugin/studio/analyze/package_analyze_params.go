package analyze

import (
	"net/url"

	"github.com/UiPath/uipathcli/auth"
)

type packageAnalyzeParams struct {
	Organization          string
	Tenant                string
	BaseUri               url.URL
	AuthToken             *auth.AuthToken
	IdentityUri           url.URL
	Source                string
	StopOnRuleViolation   bool
	TreatWarningsAsErrors bool
	GovernanceFile        string
}

func newPackageAnalyzeParams(
	organization string,
	tenant string,
	baseUri url.URL,
	authToken *auth.AuthToken,
	identityUri url.URL,
	source string,
	stopOnRuleViolation bool,
	treatWarningsAsErrors bool,
	governanceFile string,
) *packageAnalyzeParams {
	return &packageAnalyzeParams{
		organization,
		tenant,
		baseUri,
		authToken,
		identityUri,
		source,
		stopOnRuleViolation,
		treatWarningsAsErrors,
		governanceFile,
	}
}
