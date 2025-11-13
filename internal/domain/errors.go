package domain

import "errors"

var (
	ErrTeamNotFound         = errors.New("team not found")
	ErrPullRequestNotFound  = errors.New("pull request not found")
	ErrPullRequestMerged    = errors.New("pull request merged")
	ErrReviewerAlreadyAdded = errors.New("reviewer already added")
)
