package main

import (
	"context"
	"fmt"
	"github.com/dstotijn/go-notion"
	"log/slog"
	"time"
)

const (
	WHO_CHANGE     = "Кто сделал"
	OPERATION_SUM  = "Сумма пополнения"
	OPERATION_DATE = "Дата внесения"
	COMMENT        = "Комментарий"
)

type NotionManager struct {
	log    *slog.Logger
	client *notion.Client
}

func NewNotionManager(client *notion.Client) NotionManager {
	return NotionManager{
		log:    GetLogger(),
		client: client,
	}
}

func (n *NotionManager) PrintPage(ctx context.Context, pageID string) error {
	page, err := n.client.FindPageByID(ctx, pageID)
	if err != nil {
		n.log.Error("fail to find a page by id", "err", err, "page id", pageID)
		return err
	}

	fmt.Printf("found page: %v\n", page)
	return nil
}

func (n *NotionManager) QueryDatabase() {
	databaseID := "21d3d8e126bf8022a26dc42a29697985"
	ctx := context.Background()

	query := &notion.DatabaseQuery{}
	resp, err := n.client.QueryDatabase(ctx, databaseID, query)
	if err != nil {
		n.log.Error("ошибка при запросе базы данных", "err", err)
		panic(err)
	}
	for _, p := range resp.Results {
		fmt.Printf("got page: %v\n", p)

		switch props := p.Properties.(type) {
		case notion.PageProperties:
			for _, rt := range props.Title.Title {
				fmt.Println("Заголовок:", rt.Text.Content)
			}
		case notion.DatabasePageProperties:
			for name, prop := range props {
				fmt.Println("name", name)
				value := prop.Value()
				fmt.Println(value)
			}
		default:
			log.Error(fmt.Sprintf("неожиданный тип Properties: %T", props))
			panic(err)
		}

		// Пример чтения свойства Title
		//p.UnmarshalJSON()
		//titleProp := jsonToMap(p.Properties)

		//if len(titleProp) > 0 {/
		//	log.Println("Страница:", titleProp[0].PlainText)
	}
}

func (n *NotionManager) InsertToDb(ctx context.Context) {
	databaseID := "21d3d8e126bf8022a26dc42a29697985"

	props := notion.DatabasePageProperties{
		WHO_CHANGE: notion.DatabasePageProperty{
			Type: notion.DBPropTypeTitle,
			Title: []notion.RichText{
				{
					Text: &notion.Text{Content: "Миша"},
				},
			},
		},
		OPERATION_SUM: notion.DatabasePageProperty{
			Type:   notion.DBPropTypeNumber,
			Number: notion.Float64Ptr(32918),
		},
		OPERATION_DATE: notion.DatabasePageProperty{
			Type: notion.DBPropTypeDate,
			Date: &notion.Date{
				Start: notion.NewDateTime(time.Now(), true),
				//Start: notion.TimePtr(time.Date(2025, 6, 25, 0, 0, 0, 0, time.UTC)),
			},
		},
		COMMENT: notion.DatabasePageProperty{
			Type: notion.DBPropTypeRichText,
			RichText: []notion.RichText{
				{
					Text: &notion.Text{Content: "на июнь месяц"},
				},
			},
		},
	}
	params := notion.CreatePageParams{
		ParentType:             notion.ParentTypeDatabase,
		ParentID:               databaseID,
		DatabasePageProperties: &props,
	}

	page, err := n.client.CreatePage(ctx, params)
	if err != nil {
		log.Error(fmt.Sprintf("Ошибка создания страницы: %v", err))
		return
	}
	log.Info(fmt.Sprintf("Страница создана: %s", page.URL))
}
