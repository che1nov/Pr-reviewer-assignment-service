package dto

type DeactivateTeamUsersInput struct {
	TeamName string `json:"team_name" validate:"required"`
}

type DeactivateTeamUsersOutput struct {
	DeactivatedCount  int `json:"deactivated_count"`
	ReassignedPRCount int `json:"reassigned_pr_count"`
}
