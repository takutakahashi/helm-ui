package model

type RegistryMapping struct {
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"releaseName"`
	ChartName   string `json:"chartName"`
	Registry    string `json:"registry"`
}

type SetRegistryRequest struct {
	Registry string `json:"registry" validate:"required"`
}
