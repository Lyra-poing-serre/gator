package aggregator

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Lyra-poing-serre/gator/internal/database"

	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}
	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RSSFeed{}, err
	}

	r, err := createRssFeed(data)
	if err != nil {
		return &RSSFeed{}, err
	}
	return &r, nil
}

func createRssFeed(data []byte) (RSSFeed, error) {
	var r RSSFeed
	err := xml.Unmarshal(data, &r)
	if err != nil {
		return RSSFeed{}, err
	}
	r.Channel.Title = html.UnescapeString(r.Channel.Title)
	r.Channel.Description = html.UnescapeString(r.Channel.Description)
	for i, item := range r.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		r.Channel.Item[i] = item
	}
	return r, nil
}

func scrapeFeed(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Fatalln(err)
		return
	}
	fmt.Printf("------------ NEW FEED %s------------\n", feed.Url)
	rss, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Fatalln(err)
		return
	}
	s.db.MarkFeedFetched(context.Background(), feed.ID)

	// Parse the timestamps

	for _, item := range rss.Channel.Item {
		pubDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
		if err != nil {
			fmt.Printf("Error parsing time: %v\n", err)
			continue
		}
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       stringToNullString(item.Title),
			Url:         item.Link,
			Description: stringToNullString(item.Description),
			PublishedAt: pubDate,
			FeedID:      feed.ID,
		})
		if err != nil && err.Error() != `pq: duplicate key value violates unique constraint "posts_url_key"` {
			fmt.Println(err)
		}
	}
}

func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
