package main

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"strconv"
	"strings"
	"sync"
)

var (
	errUnknownCommand = errors.New("get unknown command")
	errMessageIsNil   = errors.New("message is nil")
)

const (
	commandStart  = "start"
	commandHelp   = "help"
	commandReset  = "reset"
	commandGetWeb = "notion"
)
const linkToNotion = "https://vivid-expansion-951.notion.site/21d3d8e126bf8022a26dc42a29697985?v=21d3d8e126bf80e1a2c4000c7cbdba02&pvs=74"

type TelegramManager struct {
	numWorkers   int
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
	log          *slog.Logger

	storage Storage
	manager NotionManager
	actions map[string]ButtonAction
}

func NewTelegramManager(bot *tgbotapi.BotAPI, offset, timeout, numWorkers int, debug bool, storage Storage, manager NotionManager) *TelegramManager {
	bot.Debug = debug

	tgConfig := tgbotapi.NewUpdate(offset)
	tgConfig.Timeout = timeout

	mapButtonActions := map[string]ButtonAction{
		buttonRefill:         RefillButtonAction{storage: storage},
		buttonRemove:         RemoveButtonAction{storage: storage},
		buttonBackToMain:     ReturnToMainButtonAction{storage: storage},
		buttonLeftNote:       LeftNoteButtonAction{storage: storage},
		buttonChooseCategory: ChooseCategoryButtonAction{storage: storage},
		buttonSubmitData: SubmitButtonAction{
			storage:        storage,
			notionDatabase: manager,
		},
	}

	for _, c := range categories {
		mapButtonActions[c] = CategoryDelegate{
			storage:      storage,
			categoryName: c,
		}
	}

	tm := &TelegramManager{
		bot:        bot,
		log:        GetLogger(),
		numWorkers: numWorkers,
		storage:    storage,
		actions:    mapButtonActions,
	}

	err := tm.setCommands()
	if err != nil {
		log.Error("fail to set commands for telegram bot", "error", err)
		panic(fmt.Errorf("fail to set commands for telegram bot: %e", err))
	}
	return tm
}

func (tm *TelegramManager) ListenAndServe(ctx context.Context) error {
	upChan := tm.bot.GetUpdatesChan(tm.updateConfig)

	var wg sync.WaitGroup
	wg.Add(tm.numWorkers)

	for i := 0; i < tm.numWorkers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					tm.log.Warn("context for telegram is done")
					return
				case update, ok := <-upChan:
					if !ok {
						tm.log.Error("updates channel is closed")
						return
					}

					if err := tm.createUserIfMissed(update); err != nil {
						tm.log.Error("fail to create user", "error", err)
						continue
					}

					if update.CallbackQuery != nil {
						if err := tm.processCallbackQuery(update); err != nil {
							tm.log.Error("fail to process callback query", "error", err)
						}
					}

					if update.Message == nil {
						continue
					}

					if err := tm.processMessage(update); err != nil {
						if errors.Is(err, errUnknownCommand) {
							continue
						}
						tm.log.Error("fail to process command", "error", err, "user_id", update.Message.From.ID)
					}

				}
			}
		}()
	}

	wg.Wait()
	return nil
}

func (tm *TelegramManager) setCommands() error {
	cfg := tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{
		Command:     commandStart,
		Description: "Начать работу с ботом",
	}, tgbotapi.BotCommand{
		Command:     commandHelp,
		Description: "Запросить описание взаимодействия с ботом",
	}, tgbotapi.BotCommand{
		Command:     commandReset,
		Description: "Сбрасывает состояние заполнение данных для операции",
	}, tgbotapi.BotCommand{
		Command:     commandGetWeb,
		Description: "Возвращает ссылку на notion"})

	_, err := tm.bot.Request(cfg)
	return err
}

func (tm *TelegramManager) processMessage(update tgbotapi.Update) error {
	msg := update.Message
	userID := update.Message.From.ID

	if msg.IsCommand() {
		return tm.processCommand(update)
	}

	userState, err := tm.storage.Get(userID)
	if err != nil {
		return err
	}

	if !userState.isWaitUserInput {
		tm.log.Info("message is not command", "user_id", userState, "text", msg.Text)
		response := tgbotapi.NewMessage(update.Message.Chat.ID, "все хиханьки хаханьки тебе? Жмякни по /start")
		if _, err := tm.bot.Send(response); err != nil {
			return err
		}
		return nil
	}

	if userState.isWaitUserInput {
		userInput := msg.Text
		tm.log.Info("got expected input from user", "user_id", update.Message.From.ID, "userInput", userInput)

		switch userState.UserStateCurrentOperation {
		case settingSum:
			sum, err := strconv.ParseFloat(strings.TrimSpace(userInput), 64)
			if err != nil {
				_ = tm.sendSpecificErrMsg("не удалось прочесть сумму, пожалуйста введите в формате 0.321", update)
				return fmt.Errorf("fail to parse sum value: %w", err)
			}

			if sum < 0 {
				tm.log.Warn("user attempt to enter negative sum")
				_ = tm.sendSpecificErrMsg("сумма должна быть положительным числом", update)
				return errors.New("user attempt to enter negative sum")
			}

			//todo simplify?
			if userState.Status == OperationStatusRefill {
				userState.UserStateCurrentOperation = choosingNote
				if err = tm.storage.Save(userID, userState); err != nil {
					return err
				}
				button, ok := tm.actions[buttonLeftNote]
				if !ok {
					return fmt.Errorf("internal error: no action found for left note")
				}

				buttonResponse, err := button.Action(update.Message.From.ID, update)
				if err != nil {
					return err
				}
				if _, err := tm.bot.Send(buttonResponse); err != nil {
					tm.log.Error("fail to send button response", "error", err)
					return err
				}
			} else {
				userState.UserStateCurrentOperation = choosingCategory
				if err = tm.storage.Save(userID, userState); err != nil {
					return err
				}
				button, ok := tm.actions[buttonChooseCategory]
				if !ok {
					return fmt.Errorf("internal error: no action found for buttonChooseCategory")
				}
				buttonResponse, err := button.Action(update.Message.From.ID, update)
				if err != nil {
					return err
				}
				if _, err := tm.bot.Send(buttonResponse); err != nil {
					tm.log.Error("fail to send button response", "error", err)
					return err
				}
			}
			userState.OperationSum = sum
			if err := tm.storage.Save(update.Message.From.ID, userState); err != nil {
				tm.log.Error("fail to save user state", "error", err)
				return err
			}
			tm.log.Info("user entered sum", "user_id", update.Message.From.ID, "updated user state", userState)
		case choosingNote:
			button, ok := tm.actions[buttonSubmitData]
			if !ok {
				return fmt.Errorf("internal error: no action found for buttonSubmitData")
			}

			buttonResponse, err := button.Action(update.Message.From.ID, update)
			if err != nil {
				return err
			}
			if _, err := tm.bot.Send(buttonResponse); err != nil {
				tm.log.Error("fail to send button response", "error", err)
			}
		}
	}

	return nil
}

func (tm *TelegramManager) processCommand(update tgbotapi.Update) error {
	msg := update.Message

	switch msg.Command() {
	case commandStart:
		tm.log.Info("Received /start command", "user_id", msg.From.ID)
		keyboard := createMainInlineCommands()
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, "Выберите действие")
		responseMsg.ReplyMarkup = keyboard
		_, err := tm.bot.Send(responseMsg)
		if err != nil {
			return err
		}
		return nil
	case commandHelp:
		tm.log.Info("Received /help command", "user_id", msg.From.ID)
		helpText := "Здесь будет текст для команды /help"
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, helpText)
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}

	case commandReset:
		tm.log.Info("Received /reset command", "user_id", msg.From.ID)
		if err := tm.storage.Reset(update.Message.Chat.ID); err != nil {
			return err
		}
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, "Статус заполнения данных для операции сброшен")
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}
	case commandGetWeb:
		tm.log.Info("Received /get web command", "user_id", msg.From.ID)
		btn := tgbotapi.NewInlineKeyboardButtonURL("Открыть в Notion", linkToNotion)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btn),
		)
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, "Нажмите кнопку ниже, чтобы перейти в Notion:")
		responseMsg.ReplyMarkup = keyboard

		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}

	default:
		tm.log.Warn("Received unknown command", "command", msg.Command(), "user_id", msg.From.ID)
		responseMsg := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Неизвестная команда: %s", msg.Command()))
		if _, err := tm.bot.Send(responseMsg); err != nil {
			return err
		}
		return errUnknownCommand
	}
	return nil
}

func (tm *TelegramManager) processCallbackQuery(update tgbotapi.Update) error {
	msg := update.CallbackQuery
	userID := msg.From.ID

	callbackData := update.CallbackQuery.Data
	button, ok := tm.actions[callbackData]

	if !ok {
		return nil
	}

	resp, err := button.Action(userID, update)
	if err != nil {
		return err
	}
	if _, err = tm.bot.Send(resp); err != nil {
		return err
	}
	tm.log.Info("sent message to user", "user_id", userID, "command", callbackData)

	return nil
}

func (tm *TelegramManager) sendSpecificErrMsg(msgText string, update tgbotapi.Update) error {
	msg := update.Message
	userID := msg.From.ID

	errResp := tgbotapi.NewMessage(userID, msgText)
	if err := tm.storage.Reset(userID); err != nil {
		tm.log.Error("fail to reset user before sending error msg", "user_id", userID, "error", err)
		return err
	}

	if _, err := tm.bot.Send(errResp); err != nil {
		tm.log.Error("fail to send error msg", "user_id", userID, "error", err)
		return err
	}
	return nil
}

func (tm *TelegramManager) sendUnexpectedErrMsg(update tgbotapi.Update) error {
	return tm.sendSpecificErrMsg("произошла ошибка при обработке ваших ответов и/или внутренняя\nначните заново", update)
}

func (tm *TelegramManager) createUserIfMissed(update tgbotapi.Update) error {
	var chatID int64

	msg := update.Message
	callbackQuery := update.CallbackQuery
	if msg != nil {
		chatID = msg.Chat.ID
	}
	if callbackQuery != nil {
		chatID = callbackQuery.From.ID
	}

	_, err := tm.storage.Get(chatID)
	if err == nil {
		return nil
	}

	if !errors.Is(err, errUserNotFound) {
		tm.log.Error("fail to get user state", "error", err)
		return err
	}
	if err := tm.storage.Save(chatID, UserState{}); err != nil {
		return err
	}
	tm.log.Info("missed user created state", "chat_id", chatID)
	if err := tm.storage.Save(chatID, UserState{}); err != nil {
		tm.log.Error("fail to save user state", "error", err)
		return err
	}

	return nil
}
