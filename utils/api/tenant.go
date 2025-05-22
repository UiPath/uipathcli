package api

type Tenant struct {
	Id   string
	Name string
}

func NewTenant(id string, name string) *Tenant {
	return &Tenant{id, name}
}
