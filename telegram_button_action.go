package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const choseAction = "выберите действие"

const (
	buttonRefill = "пополнить"
	buttonRemove = "снятие"

	buttonBackToMain = "вернуться к стартовой странице"
)

type ButtonAction interface {
	Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error)
}

type RefillButtonAction struct{ storage Storage }

func (r RefillButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	state, err := r.storage.Get(chatID)
	if err != nil {
		return nil, err
	}
	state.isWaitUserInput = true
	state.Status = buttonRefill
	state.UserStateCurrentOperation = settingSum
	if err = r.storage.Save(chatID, state); err != nil {
		return nil, err
	}

	keys := createRefillKeys()

	msgID := update.CallbackQuery.Message.MessageID
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, "Введите сумму операции", keys)

	return response, nil
}

type RemoveButtonAction struct{ storage Storage }

func (r RemoveButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	state, err := r.storage.Get(chatID)
	if err != nil {
		return nil, err
	}
	state.isWaitUserInput = true
	state.Status = buttonRemove
	state.UserStateCurrentOperation = settingSum
	if err = r.storage.Save(chatID, state); err != nil {
		return nil, err
	}

	keys := createRemoveKeys()
	msgID := update.CallbackQuery.Message.MessageID
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, choseAction, keys)
	return response, nil
}

type ReturnToMainButtonAction struct{ storage Storage }

func (r ReturnToMainButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	keys := createMainInlineCommands()
	msgID := update.CallbackQuery.Message.MessageID
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, choseAction, keys)
	return response, nil
}
