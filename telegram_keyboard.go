package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func createMainInlineCommands() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonRefill, buttonRefill),
			tgbotapi.NewInlineKeyboardButtonData(buttonRemove, buttonRemove),
		),
	)
}

func createRefillKeys() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("назад", buttonBackToMain),
		),
	)
}

func createRemoveKeys() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("удоли меня", "звуки удаления"),
		), tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("назад", buttonBackToMain),
		),
	)
}
