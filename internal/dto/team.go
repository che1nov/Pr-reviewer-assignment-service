package dto

type TeamInput struct {
	Name  string      `json:"name"`
	Users []UserInput `json:"users"`
}

type TeamOutput struct {
	Name  string       `json:"name"`
	Users []UserOutput `json:"users"`
}
