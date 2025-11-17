package httpcontroller

// Коды ошибок для API
const (
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeInternal     = "INTERNAL"
	ErrCodeNoCandidate  = "NO_CANDIDATE"
	ErrCodeTeamExists   = "TEAM_EXISTS"
	ErrCodePRExists     = "PR_EXISTS"
	ErrCodePRMerged     = "PR_MERGED"
	ErrCodeNotAssigned  = "NOT_ASSIGNED"
)

// Сообщения об ошибках
const (
	ErrMsgInternalError = "internal error"
)

