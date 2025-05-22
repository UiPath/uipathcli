package api

type Organization struct {
	Id      string
	Name    string
	Tenants []Tenant
}

func NewOrganization(id string, name string, tenants []Tenant) *Organization {
	return &Organization{id, name, tenants}
}
