package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	buttonRefill = "пополнить"
	buttonRemove = "снятие"

	buttonBackToMain = "вернуться к стартовой странице"
)

type ButtonAction interface {
	Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error)
}

type RefillButtonAction struct{}

func (r RefillButtonAction) Action(chatID int64, _ tgbotapi.Update) (tgbotapi.Chattable, error) {
	keyboard := createRefillKeys()
	response := tgbotapi.NewMessage(chatID, "ты шо ебанутый?")
	response.ReplyMarkup = keyboard

	return response, nil
}

type RemoveButtonAction struct{}

func (r RemoveButtonAction) Action(chatID int64, _ tgbotapi.Update) (tgbotapi.Chattable, error) {
	keyboard := createRemoveKeys()
	response := tgbotapi.NewMessage(chatID, "Шо")
	response.ReplyMarkup = keyboard
	return response, nil
}
