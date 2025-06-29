package main

import (
	"context"
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
