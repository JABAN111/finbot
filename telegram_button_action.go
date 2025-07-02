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

type RefillButtonAction struct{}

func (r RefillButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	keys := createRefillKeys()
	msgID := update.CallbackQuery.Message.MessageID
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, choseAction, keys)

	return response, nil
}

type RemoveButtonAction struct{}

func (r RemoveButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	keys := createRemoveKeys()
	msgID := update.CallbackQuery.Message.MessageID
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, choseAction, keys)
	return response, nil
}

type ReturnToMainButtonAction struct{}

func (r ReturnToMainButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	keys := createMainInlineCommands()
	msgID := update.CallbackQuery.Message.MessageID
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, choseAction, keys)
	return response, nil
}
