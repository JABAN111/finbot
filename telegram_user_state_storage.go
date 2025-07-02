package main

import (
	"errors"
	"log/slog"
	"sync"
)

type Storage interface {
	Save(chatID int64, state UserState) error
	Get(chatID int64) (UserState, error)
	Reset(chatID int64) error
}

const (
	beginOperation = iota
	settingSum
	choosingStatus
	choosingCategory
	choosingComment
)

var (
	errUserNotFound = errors.New("user not found")
)

type UserState struct {
	UserStateCurrentOperation int
	isWaitUserInput           bool

	OperationSum float64
	Category     string
	Status       OperationStatus
	Comment      string
}

type UserStateStorage struct {
	mu    sync.Mutex
	users map[int64]UserState
	log   *slog.Logger
}

func NewUserStateStorage() *UserStateStorage {
	return &UserStateStorage{
		users: make(map[int64]UserState),
		log:   GetLogger(),
	}
}

func (u *UserStateStorage) Save(chatID int64, state UserState) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.users[chatID] = state
	u.log.Info("save user to user state storage", "chatID", chatID, "userState", state)
	return nil
}

func (u *UserStateStorage) Get(chatID int64) (UserState, error) {
	state, ok := u.users[chatID]
	if !ok {
		u.log.Warn("user state not found", "chatID", chatID)
		return UserState{}, errUserNotFound
	}
	return state, nil
}

func (u *UserStateStorage) Reset(chatID int64) error {
	_, err := u.Get(chatID)
	if err != nil {
		return err
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.users[chatID] = UserState{
		UserStateCurrentOperation: beginOperation,
		isWaitUserInput:           false,
		OperationSum:              0,
		Category:                  "",
		Comment:                   "",
	}
	u.log.Info("reset user to user state storage", "chatID", chatID)

	return nil
}
