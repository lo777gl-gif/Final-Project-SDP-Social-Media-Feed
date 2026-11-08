package adapter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"final/internal/structs"
)

// Target: every external source must give us normal posts.
type ContentSource interface {
	Name() string
	FetchPosts(limit int) ([]structs.Post, error)
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Adaptee 1: JSON API with its own post format.
type JSONPlaceholderPost struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type JSONPlaceholderClient struct {
	URL string
}

func (c *JSONPlaceholderClient) GetPosts() ([]JSONPlaceholderPost, error) {
	resp, err := httpClient.Get(c.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("json source returned %d", resp.StatusCode)
	}
	var posts []JSONPlaceholderPost
	err = json.NewDecoder(resp.Body).Decode(&posts)
	return posts, err
}

// Adaptee 2: RSS feed with its own item format.
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type rssDoc struct {
	Items []RSSItem `xml:"channel>item"`
}

type RSSClient struct {
	URL string
}

func (c *RSSClient) GetItems() ([]RSSItem, error) {
	resp, err := httpClient.Get(c.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rss source returned %d", resp.StatusCode)
	}
	var doc rssDoc
	err = xml.NewDecoder(resp.Body).Decode(&doc)
	return doc.Items, err
}

// Adapter 1: JSON API -> Post
type JSONPlaceholderAdapter struct {
	Client *JSONPlaceholderClient
}

func NewJSONPlaceholderAdapter(url string) *JSONPlaceholderAdapter {
	return &JSONPlaceholderAdapter{Client: &JSONPlaceholderClient{URL: url}}
}

func (a *JSONPlaceholderAdapter) Name() string {
	return "jsonplaceholder"
}

func (a *JSONPlaceholderAdapter) FetchPosts(limit int) ([]structs.Post, error) {
	items, err := a.Client.GetPosts()
	if err != nil {
		return nil, err
	}
	posts := []structs.Post{}
	for _, it := range items {
		if len(posts) == limit {
			break
		}
		posts = append(posts, structs.Post{
			ID:             it.ID,
			UserID:         it.UserID,
			AuthorUsername: fmt.Sprintf("jsonplaceholder user %d", it.UserID),
			Kind:           structs.PostKindText,
			Title:          it.Title,
			Content:        it.Body,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		})
	}
	return posts, nil
}

// Adapter 2: RSS -> Post
type RSSAdapter struct {
	SourceName string
	Client     *RSSClient
}

func NewRSSAdapter(name, url string) *RSSAdapter {
	return &RSSAdapter{SourceName: name, Client: &RSSClient{URL: url}}
}

func (a *RSSAdapter) Name() string {
	return a.SourceName
}

func (a *RSSAdapter) FetchPosts(limit int) ([]structs.Post, error) {
	items, err := a.Client.GetItems()
	if err != nil {
		return nil, err
	}
	posts := []structs.Post{}
	for i, it := range items {
		if len(posts) == limit {
			break
		}
		created, err := time.Parse(time.RFC1123Z, it.PubDate)
		if err != nil {
			created, err = time.Parse(time.RFC1123, it.PubDate)
		}
		if err != nil {
			created = time.Now().UTC()
		}
		posts = append(posts, structs.Post{
			ID:             i + 1,
			AuthorUsername: a.SourceName,
			Kind:           structs.PostKindLink,
			Title:          strings.TrimSpace(it.Title),
			Content:        strings.TrimSpace(it.Description),
			MediaURL:       strings.TrimSpace(it.Link),
			CreatedAt:      created,
			UpdatedAt:      created,
		})
	}
	return posts, nil
}
