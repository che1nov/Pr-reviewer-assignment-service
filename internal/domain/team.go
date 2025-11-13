package domain

type Team struct {
	Name  string
	Users []User
}

func NewTeam(name string, users []User) Team {
	prepared := make([]User, 0, len(users))
	for _, user := range users {
		user.TeamName = name
		prepared = append(prepared, user)
	}
	return Team{
		Name:  name,
		Users: prepared,
	}
}

func (t Team) ActiveReviewersExcluding(authorID string) []User {
	reviewers := make([]User, 0, len(t.Users))
	for _, user := range t.Users {
		if user.ID == authorID || !user.IsActive {
			continue
		}
		reviewers = append(reviewers, user)
	}
	return reviewers
}
