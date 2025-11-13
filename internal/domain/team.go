package domain

type Team struct {
	Name  string
	Users []User
}

func NewTeam(name string, users []User) Team {
	return Team{
		Name:  name,
		Users: users,
	}
}
