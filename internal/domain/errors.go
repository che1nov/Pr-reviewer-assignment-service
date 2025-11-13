package domain

import "errors"

var (
	ErrTeamExists           = errors.New("команда уже существует")
	ErrTeamNotFound         = errors.New("команда не найдена")
	ErrPullRequestExists    = errors.New("pull request уже существует")
	ErrPullRequestNotFound  = errors.New("pull request не найден")
	ErrPullRequestMerged    = errors.New("pull request в статусе merged")
	ErrNoReviewerCandidates = errors.New("нет доступных кандидатов в ревьюеры")
	ErrReviewerAlreadyAdded = errors.New("ревьюер уже назначен")
	ErrReviewerInactive     = errors.New("ревьюер неактивен")
	ErrReviewerNotInTeam    = errors.New("ревьюер не в команде автора")
	ErrReviewerIsAuthor     = errors.New("автор не может быть ревьюером")
	ErrReviewerLimitReached = errors.New("достигнут лимит ревьюеров")
	ErrReviewerNotAssigned  = errors.New("ревьюер не назначен")
	ErrUserNotFound         = errors.New("пользователь не найден")
)
