package domain

type PullRequest struct {
	ID                string
	Title             string
	AuthorID          string
	TeamName          string
	Reviewers         []string
	Status            string
	NeedMoreReviewers bool
}

func NewPullRequest(id, title, authorID, teamName string) PullRequest {
	return PullRequest{
		ID:                id,
		Title:             title,
		AuthorID:          authorID,
		TeamName:          teamName,
		Status:            "OPEN",
		Reviewers:         []string{},
		NeedMoreReviewers: true,
	}
}

func (pr *PullRequest) AssignReviewers(reviewers []string) {
	pr.Reviewers = reviewers
	pr.NeedMoreReviewers = len(reviewers) < 2
}
