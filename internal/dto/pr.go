package dto

type CreatePullRequestInput struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	AuthorID string `json:"author_id"`
	TeamName string `json:"team_name"`
}

type PullRequestOutput struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	AuthorID          string   `json:"author_id"`
	TeamName          string   `json:"team_name"`
	Reviewers         []string `json:"reviewers"`
	Status            string   `json:"status"`
	NeedMoreReviewers bool     `json:"need_more_reviewers"`
}

type ReviewerPullRequestsOutput struct {
	ReviewerID   string               `json:"reviewer_id"`
	PullRequests []PullRequestOutput `json:"pull_requests"`
}
