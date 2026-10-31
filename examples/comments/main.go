// Command comments lists an event's comments or retrieves one comment thread.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cagelin/polymarket-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	thread := flag.String("thread", "", "retrieve one comment thread by comment ID instead of an event list")
	pageSize := flag.Int("limit", 10, "comments in the first list page (1-100)")
	flag.Parse()
	listMode := *thread == "" && flag.NArg() == 1
	threadMode := *thread != "" && flag.NArg() == 0
	if (!listMode && !threadMode) || *pageSize < 1 || *pageSize > 100 {
		return fmt.Errorf("usage: go run ./examples/comments [-limit 10] <event-id> OR go run ./examples/comments -thread <comment-id>")
	}
	client, err := polymarket.NewPublicClient(polymarket.ClientConfig{})
	if err != nil {
		return fmt.Errorf("create public client: %w", err)
	}
	defer func() { _ = client.Close() }() // Close only releases local resources.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	encoder := json.NewEncoder(os.Stdout)
	if threadMode {
		comments, threadErr := client.GetCommentThread(ctx, polymarket.CommentID(*thread), nil)
		if threadErr != nil {
			return fmt.Errorf("read comment thread: %w", threadErr)
		}
		if err = encoder.Encode(comments); err != nil {
			return fmt.Errorf("write comment thread: %w", err)
		}
		return nil
	}
	pager, err := client.ListComments(polymarket.ListCommentsParams{
		ParentEntityID: flag.Arg(0), ParentEntityType: polymarket.CommentParentEvent, PageSize: pageSize,
	})
	if err != nil {
		return fmt.Errorf("select event comments: %w", err)
	}
	page, err := pager.FirstPage(ctx)
	if err != nil {
		return fmt.Errorf("read event comments: %w", err)
	}
	if err = encoder.Encode(page); err != nil {
		return fmt.Errorf("write event comments: %w", err)
	}
	if page.LimitReached {
		return fmt.Errorf("service pagination limit reached; completeness is unknown")
	}
	return nil
}
