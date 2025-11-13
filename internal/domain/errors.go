package domain

import "errors"

var (
	ErrTeamNotFound        = errors.New("team not found")
	ErrPullRequestNotFound = errors.New("pull request not found")
)
