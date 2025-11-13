package domain

type User struct {
	ID       string `db:"id"`
	Name     string `db:"name"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
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
