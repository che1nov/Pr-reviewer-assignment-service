package dto

type CreatePullRequestInput struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	AuthorID  string   `json:"author_id"`
	TeamName  string   `json:"team_name"`
	Reviewers []string `json:"reviewers"`
}

type PullRequestOutput struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	AuthorID  string   `json:"author_id"`
	TeamName  string   `json:"team_name"`
	Reviewers []string `json:"reviewers"`
	Status    string   `json:"status"`
}
