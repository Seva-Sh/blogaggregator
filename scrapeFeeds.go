package main

import (
	"context"
	"database/sql"
	"fmt"
	"gator/internal/database"
	"strings"
	"time"

	"github.com/google/uuid"
)

func scrapeFeeds(s *state) error {
	ctx := context.Background()
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		fmt.Println("failed to fetch feed")
		return err
	}

	markedFeed, err := s.db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		fmt.Println("failed to mark feed fetched")
		return err
	}

	fetchedFeed, err := fetchFeed(ctx, markedFeed.Url)
	if err != nil {
		fmt.Println("failed to fetch the feed")
		return err
	}

	for _, item := range fetchedFeed.Channel.Item {
		parsedTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		publishedAt := sql.NullTime{}
		if err == nil {
			publishedAt = sql.NullTime{Time: parsedTime, Valid: true}
		} else {
			fmt.Println("failed to parse time")
		}
		post, err := s.db.CreatePost(ctx, database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  item.Description != "",
			},
			PublishedAt: publishedAt,
			FeedID:      markedFeed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				// skip duplicate URL
				continue
			}
			fmt.Println("failed to create post:", err)
			return err
		}
		fmt.Println("created post: %s", post.Title)
	}

	return nil
}
