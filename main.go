package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dapoadedire/hackernews-daily-bot/controller"
	"github.com/dapoadedire/hackernews-daily-bot/database"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

type Article struct {
	Title string
	Link  string
}

type Category struct {
	Name     string
	Articles []Article
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

func getArticles(url string, limit int) ([]Article, error) {
	c := colly.NewCollector(
		colly.AllowedDomains("news.ycombinator.com"),
		colly.MaxDepth(1),
	)

	c.SetRequestTimeout(30 * time.Second)
	var articles []Article
	var mu sync.Mutex

	c.OnHTML("tr.athing", func(e *colly.HTMLElement) {
		mu.Lock()
		defer mu.Unlock()

		if len(articles) >= limit {
			return
		}

		title := e.ChildText("td.title > span.titleline > a")
		link := e.ChildAttr("td.title > span.titleline > a", "href")

		if title != "" && link != "" {
			articles = append(articles, Article{
				Title: title,
				Link:  fmt.Sprintf("https://news.ycombinator.com/%s", link),
			})
		}
	})

	err := c.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch articles: %w", err)
	}

	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles found")
	}

	return articles, nil
}

func formatMessage(username string, categories []Category) string {
	messageText := fmt.Sprintf("Good morning, %s!\n\nHere are the Hacker News articles that match your keywords for today:\n\n", username)

	for _, category := range categories {
		messageText += fmt.Sprintf("%s\n", category.Name)
		for i, article := range category.Articles {
			messageText += fmt.Sprintf("%d. [%s](%s)\n", i+1, article.Title, article.Link)
		}
		messageText += "\n"
	}

	messageText += "Stay curious and enjoy reading! 🚀\n\nBest,\nHackerNews Daily Bot"
	return messageText
}

func sendMessage(userID int64, username string, categories []Category) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	messageText := formatMessage(username, categories)
	data := url.Values{
		"chat_id":    {strconv.FormatInt(userID, 10)},
		"text":       {messageText},
		"parse_mode": {"Markdown"},
	}

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func sendToAllUsers(ctx context.Context, categories []Category) error {
	users, err := controller.GetUsers()
	if err != nil {
		return fmt.Errorf("error fetching users: %w", err)
	}
	log.Printf("Fetched %d users", len(users))

	errChan := make(chan error, len(users))
	sem := make(chan struct{}, 10) // Limit concurrent goroutines to 10

	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(user controller.User) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}

			userID, err := strconv.ParseInt(user.UserID, 10, 64)
			if err != nil {
				errChan <- fmt.Errorf("error converting userID for %s: %w", user.Username, err)
				return
			}

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				if err := sendMessage(userID, user.Username, categories); err != nil {
					errChan <- fmt.Errorf("error sending message to %s: %w", user.Username, err)
					return
				}
			}

			log.Printf("Message sent successfully to %s", user.Username)
		}(user)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errors []error
	for err := range errChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while sending messages: %v", errors)
	}

	return nil
}

func fetchCategories() ([]Category, error) {
	categories := []struct {
		name string
		url  string
	}{
		{"Ask HN", "https://news.ycombinator.com/ask"},
		{"Show HN", "https://news.ycombinator.com/show"},
		{"Top News", "https://news.ycombinator.com/news"},
	}

	var results []Category
	for _, cat := range categories {
		articles, err := getArticles(cat.url, 5)
		if err != nil {
			return nil, fmt.Errorf("error fetching %s articles: %w", cat.name, err)
		}
		results = append(results, Category{Name: cat.name, Articles: articles})
	}

	return results, nil
}

func main() {
	timeStart := time.Now()
	database.InitDB()
	defer database.DB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	categories, err := fetchCategories()
	if err != nil {
		log.Fatalf("Error fetching categories: %v", err)
	}
	log.Printf("Fetched articles for %d categories", len(categories))

	if err := sendToAllUsers(ctx, categories); err != nil {
		log.Fatalf("Error sending messages: %v", err)
	}

	log.Println("Messages sent successfully to all users")
	log.Printf("Execution time: %v", time.Since(timeStart))
}
