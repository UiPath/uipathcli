package studio

type studioProjectJson struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	ProjectId       string `json:"projectId"`
	TargetFramework string `json:"targetFramework"`
}
