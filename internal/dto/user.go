package dto

type UserInput struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive *bool  `json:"is_active,omitempty"`
}

type UserOutput struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type UsersOutput struct {
	Users []UserOutput `json:"users"`
}
