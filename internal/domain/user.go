package domain

type User struct {
	ID       string
	Name     string
	TeamName string
	IsActive bool
}

// NewUser создаёт пользователя с привязкой к команде.
func NewUser(id, name, teamName string, isActive bool) User {
	return User{
		ID:       id,
		Name:     name,
		TeamName: teamName,
		IsActive: isActive,
	}
}
