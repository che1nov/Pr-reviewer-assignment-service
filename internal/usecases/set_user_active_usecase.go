package usecases

import (
	"context"
	"log/slog"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

// SetUserActiveUseCase обновляет признак активности пользователя.
type SetUserActiveUseCase struct {
	users UserStorage
	log   *slog.Logger
}

// NewSetUserActiveUseCase создаёт use case установки активности пользователя.
func NewSetUserActiveUseCase(storage UserStorage, log *slog.Logger) *SetUserActiveUseCase {
	return &SetUserActiveUseCase{
		users: storage,
		log:   log,
	}
}

// Execute включает или выключает пользователя.
func (uc *SetUserActiveUseCase) Execute(ctx context.Context, id string, isActive bool) (domain.User, error) {
	uc.log.InfoContext(ctx, "изменяем активность пользователя", "user_id", id, "is_active", isActive)

	user, err := uc.users.GetUser(ctx, id)
	if err != nil {
		uc.log.WarnContext(ctx, "пользователь не найден", "user_id", id, "error", err)
		return domain.User{}, err
	}

	user.IsActive = isActive
	if err := uc.users.UpdateUser(ctx, user); err != nil {
		uc.log.ErrorContext(ctx, "не удалось обновить пользователя", "user_id", id, "error", err)
		return domain.User{}, err
	}

	uc.log.InfoContext(ctx, "статус пользователя изменён", "user_id", id, "is_active", user.IsActive)
	return user, nil
}
