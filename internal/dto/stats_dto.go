package dto

type UserStats struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	TeamName    string `json:"team_name"`
	AssignedPRs int    `json:"assigned_prs_count"`
	OpenPRs     int    `json:"open_prs_count"`
	MergedPRs   int    `json:"merged_prs_count"`
}

type PRStats struct {
	TotalPRs       int `json:"total_prs"`
	OpenPRs        int `json:"open_prs"`
	MergedPRs      int `json:"merged_prs"`
	ReviewersCount int `json:"total_reviewers"`
}

type StatsResponse struct {
	PRStats   PRStats     `json:"pr_stats"`
	UserStats []UserStats `json:"user_stats"`
}
