package dto

type UserInput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
