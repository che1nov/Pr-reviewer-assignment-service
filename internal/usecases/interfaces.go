package usecases

import "context"

type UserStorage interface {
	CreateUser(ctx context.Context, id, name string) error
}
