package interfaces

import (
	"context"
	"golepi-fsm/internal/domain/models"
	"slices"
)

type Storage interface {
	Get(ctx context.Context, chatID int64, userID int64) (models.State, error)

	Set(ctx context.Context, chatID int64, userID int64, state models.State) error

	Delete(ctx context.Context, chatID int64, userID int64) error

	GetData(ctx context.Context, chatID int64, userID int64, key string) (interface{}, error)

	SetData(ctx context.Context, chatID int64, userID int64, key string, value interface{}) error

	ClearData(ctx context.Context, chatID int64, userID int64) error
}


type FSMContext struct {
	storage Storage
	chatID  int64
	userID  int64
	ctx     context.Context
}

func NewFSMContext(ctx context.Context, storage Storage, chatID, userID int64) *FSMContext {
	return &FSMContext{
		storage: storage,
		chatID:  chatID,
		userID:  userID,
		ctx:     ctx,
	}
}

func (f *FSMContext) Current() (models.State, error) {
	return f.storage.Get(f.ctx, f.chatID, f.userID)
}

func (f *FSMContext) IsInState(states ...models.State) (bool, error) {
	current, err := f.Current()
	if err != nil {
		return false, err
	}

	if slices.Contains(states, current) {
		return true, nil
	}

	return false, nil
}


func (f *FSMContext) Set(state models.State) error {
	return f.storage.Set(f.ctx, f.chatID, f.userID, state)
}

func (f *FSMContext) Finish() error {
	return f.storage.Delete(f.ctx, f.chatID, f.userID)
}

func (f *FSMContext) ResetState() error {
	return f.storage.Set(f.ctx, f.chatID, f.userID, "")
}

func (f *FSMContext) GetData(key string) (any, error) {
	return f.storage.GetData(f.ctx, f.chatID, f.userID, key)
}

func (f *FSMContext) SetData(key string, value interface{}) error {
	return f.storage.SetData(f.ctx, f.chatID, f.userID, key, value)
}

func (f *FSMContext) UpdateData(key string, updateFn func(any) interface{}) error {
	value, err := f.GetData(key)
	if err != nil {
		return err
	}

	newValue := updateFn(value)
	return f.SetData(key, newValue)
}

func (f *FSMContext) ClearData() error {
	return f.storage.ClearData(f.ctx, f.chatID, f.userID)
}
