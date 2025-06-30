package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"sync"
)

type TelegramManager struct {
	numWorkers   int
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
	log          *slog.Logger
}

func NewTelegramManager(bot *tgbotapi.BotAPI, offset, timeout, numWorkers int, debug bool) TelegramManager {
	bot.Debug = debug

	tgConfig := tgbotapi.NewUpdate(offset)
	tgConfig.Timeout = timeout

	return TelegramManager{
		bot:        bot,
		log:        GetLogger(),
		numWorkers: numWorkers,
	}
}

func (tm *TelegramManager) ListenAndServe(ctx context.Context) error {
	upChan := tm.bot.GetUpdatesChan(tm.updateConfig)

	var wg sync.WaitGroup
	wg.Add(tm.numWorkers)

	for range tm.numWorkers {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					tm.log.Warn("context for telegram is done")
					return
				case data, ok := <-upChan:
					if !ok {
						return
					}

					if data.Message != nil {
						msg := tgbotapi.NewMessage(data.Message.Chat.ID, "Выберите действие:")
						msg.ReplyMarkup = createInlineKeyboard()
						if _, err := tm.bot.Send(msg); err != nil {
							tm.log.Error("failed to send message", "error", err)
						}
					}

					if data.CallbackQuery != nil {
						callback := tgbotapi.NewCallback(data.CallbackQuery.ID, "Выбрано: "+data.CallbackQuery.Data)
						if _, err := tm.bot.Send(callback); err != nil {
							tm.log.Error("failed to answer callback", "error", err)
						}

						responseMsg := tgbotapi.NewMessage(
							data.CallbackQuery.Message.Chat.ID,
							"Вы выбрали: "+data.CallbackQuery.Data,
						)
						msg, err := tm.bot.Send(responseMsg)
						if err != nil {
							tm.log.Error("failed to send response message", "error", err)
						}
						if _, err := tm.bot.Send(responseMsg); err != nil {
							tm.log.Error("failed to send response", "error", err)
						}
						fmt.Println(msg)
					}
					//tm.log.Info("got update", "data", data)
					//call := tgbotapi.NewCallback("2", "test")
					//msg, err := tm.bot.Send(call)
					//if err != nil {
					//	panic(err)
					//}
					//tm.log.Info("got message", "data", msg)
				}
			}
		}()
	}
	wg.Wait()

	return nil
}

// processMessage обрабатывает пользовательский запрос и, если необходимо, возвращает ему ответ
func processMessage(update tgbotapi.Update) (*tgbotapi.Message, error) {
	return nil, nil
	//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	//msg.ReplyToMessageID = update.Message.MessageID
	//
	//return msg, nil
}

func createInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Кнопка 1", "action_1"),
			tgbotapi.NewInlineKeyboardButtonData("Кнопка 2", "action_2"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Кнопка с параметрами", "action_with_params?id=123&type=test"),
			tgbotapi.NewInlineKeyboardButtonURL("Ссылка", "https://google.com"),
		),
	)
}
