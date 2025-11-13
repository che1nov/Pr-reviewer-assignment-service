package domain

type PullRequest struct {
	ID        string
	Title     string
	AuthorID  string
	TeamName  string
	Reviewers []string
	Status    string
}

func NewPullRequest(id, title, authorID, teamName string, reviewers []string) PullRequest {
	return PullRequest{
		ID:        id,
		Title:     title,
		AuthorID:  authorID,
		TeamName:  teamName,
		Reviewers: reviewers,
		Status:    "OPEN",
	}
}
