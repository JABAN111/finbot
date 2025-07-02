package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dstotijn/go-notion"
	"log/slog"
	"time"
)

type NotionDatabase interface {
	InsertOperation(ctx context.Context, operationDTO InsertOperationDto) error
}

type OperationStatus string

const (
	OperationStatusRefill = OperationStatus("Пополнение")
	OperationRemove       = OperationStatus("Снятие")
)

const (
	columnWho           = "Кто"
	columnOperationSum  = "Сумма операции"
	columnComment       = "Комментарий"
	columnCategory      = "Категория"
	columnStatus        = "Статус"
	columnOperationDate = "Дата внесения"
)

var (
	errMissedField            = errors.New("miss required field")
	errNegativeSum            = errors.New("sum must be positive")
	errInvalidOperationStatus = errors.New("invalid operation status")
)

type NotionManager struct {
	log        *slog.Logger
	client     *notion.Client
	databaseID string
}

func NewNotionManager(client *notion.Client, databaseID string) NotionManager {
	return NotionManager{
		log:        GetLogger(),
		databaseID: databaseID,
		client:     client,
	}
}

type InsertOperationDto struct {
	Creator  string
	Category string
	Sum      float64
	Status   OperationStatus
	Comment  string
}

func (n NotionManager) InsertOperation(ctx context.Context, operationDTO InsertOperationDto) error {
	if operationDTO.Status != OperationStatusRefill && operationDTO.Status != OperationRemove {
		return errInvalidOperationStatus
	}
	if operationDTO.Sum < 0 {
		return errNegativeSum
	}

	if len(operationDTO.Creator) == 0 {
		return fmt.Errorf("creator is missed, %e", errMissedField)
	}
	if len(operationDTO.Category) == 0 {
		return fmt.Errorf("category is missed, %e", errMissedField)
	}

	props := notion.DatabasePageProperties{
		columnWho: notion.DatabasePageProperty{
			Select: &notion.SelectOptions{
				Name: operationDTO.Creator,
			},
		},
		columnCategory: notion.DatabasePageProperty{
			Select: &notion.SelectOptions{
				Name: operationDTO.Category,
			},
		},
		columnComment: notion.DatabasePageProperty{
			Title: []notion.RichText{
				{
					Text: &notion.Text{Content: operationDTO.Comment},
				},
			},
		},
		columnOperationSum: notion.DatabasePageProperty{
			Number: &operationDTO.Sum,
		},
		columnOperationDate: notion.DatabasePageProperty{
			Date: &notion.Date{
				Start: notion.NewDateTime(time.Now(), false),
			},
		},
		columnStatus: notion.DatabasePageProperty{
			Status: &notion.SelectOptions{
				Name: string(operationDTO.Status),
			},
		},
	}

	params := notion.CreatePageParams{
		ParentType:             notion.ParentTypeDatabase,
		ParentID:               n.databaseID,
		DatabasePageProperties: &props,
	}

	page, err := n.client.CreatePage(ctx, params)
	if err != nil {
		log.Error("Ошибка создания страницы", "err", err)
		return err
	}
	log.Info("Страница создана", "url", page.URL)
	return nil
}
