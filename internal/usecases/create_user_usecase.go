package usecases

import "context"

type CreateUserUseCase struct {
	storage UserStorage
}

func NewCreateUserUseCase(storage UserStorage) *CreateUserUseCase {
	return &CreateUserUseCase{storage: storage}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, id, name string) error {
	return uc.storage.CreateUser(ctx, id, name)
}
