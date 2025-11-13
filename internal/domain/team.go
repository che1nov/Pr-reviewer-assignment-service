package domain

// Team описывает команду пользователей
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

func (t Team) ActiveReviewersExcluding(authorID string) []User {
	reviewers := make([]User, 0, len(t.Users))
	for _, user := range t.Users {
		if user.ID == authorID {
			continue
		}
		reviewers = append(reviewers, user)
	}
	return reviewers
}
