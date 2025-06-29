package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dstotijn/go-notion"
	"log/slog"
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

func jsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		panic(err)
	}
	return result
}
