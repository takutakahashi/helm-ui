package model

type Repository struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type AddRepositoryRequest struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url" validate:"required,url"`
}
