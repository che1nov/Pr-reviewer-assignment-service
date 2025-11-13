package domain

import "errors"

var (
	ErrTeamNotFound         = errors.New("team not found")
	ErrPullRequestNotFound  = errors.New("pull request not found")
	ErrPullRequestMerged    = errors.New("pull request merged")
	ErrReviewerAlreadyAdded = errors.New("reviewer already added")
	ErrReviewerInactive     = errors.New("reviewer inactive")
	ErrReviewerNotInTeam    = errors.New("reviewer not in team")
	ErrReviewerIsAuthor     = errors.New("reviewer is author")
	ErrReviewerLimitReached = errors.New("reviewer limit reached")
	ErrUserNotFound         = errors.New("user not found")
)
