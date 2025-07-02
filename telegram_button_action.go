package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const choseAction = "выберите действие"

const (
	buttonRefill         = "пополнить"
	buttonRemove         = "снятие"
	buttonChooseCategory = "выберите категорию"
	buttonLeftNote       = "оставить комментарий"

	buttonBackToMain = "вернуться к стартовой странице"

	buttonSubmitData = "сохранить запись в notion"
)

const (
	categoryProducts   = "продукты"
	categoryFun        = "развлечения"
	categoryKommunalka = "коммуналка"
)

var categories = []string{
	categoryKommunalka,
	categoryProducts,
	categoryFun,
}

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
	state.Status = OperationStatusRefill
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
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, "Введите сумму", keys)
	return response, nil
}

type ReturnToMainButtonAction struct{ storage Storage }

func (r ReturnToMainButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	keys := createMainInlineCommands()
	msgID := update.CallbackQuery.Message.MessageID
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, choseAction, keys)
	return response, nil
}

type ChooseCategoryButtonAction struct{ storage Storage }

func (c ChooseCategoryButtonAction) Action(chatID int64, _ tgbotapi.Update) (tgbotapi.Chattable, error) {

	keys := createChooseCategoryButtons(categories)

	resp := tgbotapi.NewMessage(chatID, "Выберите категорию")
	resp.ReplyMarkup = keys
	return resp, nil
}

type CategoryDelegate struct {
	storage      Storage
	categoryName string
}

func (c CategoryDelegate) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	state, err := c.storage.Get(chatID)
	if err != nil {
		return nil, err
	}
	state.Category = c.categoryName
	state.isWaitUserInput = true
	if err = c.storage.Save(chatID, state); err != nil {
		return nil, err
	}

	keys := createLeftNoteButtons()
	msgID := update.CallbackQuery.Message.MessageID
	response := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, "отправьте отдельным сообщением комментарий", keys)
	return response, nil
}

type LeftNoteButtonAction struct{ storage Storage }

func (l LeftNoteButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	userState, err := l.storage.Get(chatID)
	if err != nil {
		return nil, err
	}

	msg := update.Message.Text
	userState.Comment = msg
	if err = l.storage.Save(chatID, userState); err != nil {
		return nil, err
	}

	keys := createLeftNoteButtons()
	response := tgbotapi.NewMessage(chatID, "отправьте отдельным сообщением комментарий")
	response.ReplyMarkup = keys
	return response, nil
}

type SubmitButtonAction struct {
	storage        Storage
	notionDatabase NotionDatabase
}

func (s SubmitButtonAction) Action(chatID int64, update tgbotapi.Update) (tgbotapi.Chattable, error) {
	var username string
	if update.CallbackQuery != nil {
		username = update.CallbackQuery.From.UserName
	} else {
		username = update.Message.From.UserName
	}

	userState, err := s.storage.Get(chatID)
	if err != nil {
		return nil, err
	}

	category := userState.Category
	if userState.Status == OperationStatusRefill {
		category = "пополнение"
	}

	dto := InsertOperationDto{
		Creator:  username,
		Category: category,
		Sum:      userState.OperationSum,
		Status:   userState.Status,
		Comment:  userState.Comment,
	}

	if err = s.notionDatabase.InsertOperation(context.Background(), dto); err != nil {
		return nil, err
	}

	if err := s.storage.Reset(chatID); err != nil {
		return nil, err
	}

	responseMsg := tgbotapi.NewMessage(chatID, "Данные успешно отправлены")
	return responseMsg, nil
}
