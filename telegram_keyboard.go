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
			tgbotapi.NewInlineKeyboardButtonData("назад", buttonBackToMain),
		),
	)
}

func createChooseCategoryButtons(categoryButtons []string) tgbotapi.InlineKeyboardMarkup {
	var rows []tgbotapi.InlineKeyboardButton

	for _, button := range categoryButtons {
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(button, button))
		rows = append(rows, row...)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows)
}

func createLeftNoteButtons() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonSubmitData, buttonSubmitData),
			tgbotapi.NewInlineKeyboardButtonData(buttonBackToMain, buttonBackToMain),
		),
	)
}
