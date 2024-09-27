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

// Article represents a HackerNews article
type Article struct {
	Title string
	Link  string
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

func getArticles(url string) ([]Article, error) {
	c := colly.NewCollector(
		colly.AllowedDomains("news.ycombinator.com"),
	)

	c.SetRequestTimeout(30 * time.Second)
	var articles []Article
	var mu sync.Mutex

	c.OnHTML("tr.athing", func(e *colly.HTMLElement) {
		title := e.ChildText("td.title > span.titleline > a")
		link := e.ChildAttr("td.title > span.titleline > a", "href")

		if title != "" && link != "" {
			article := Article{
				Title: title,
				Link:  fmt.Sprintf("https://news.ycombinator.com/%s", link),
			}
			mu.Lock()
			articles = append(articles, article)
			mu.Unlock()
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

// formatMessage creates a formatted message with the articles
func formatMessage(username string, articles []Article) string {
	messageText := fmt.Sprintf("Good morning, %s!\n\nHere are the Hacker News articles that match your keywords for today:\n\n", username)

	for i, article := range articles {
		if i >= 10 {
			break
		}
		messageText += fmt.Sprintf("%d. [%s](%s)\n", i+1, article.Title, article.Link)
	}

	messageText += "\nStay curious and enjoy reading! 🚀\n\nBest,\nHackerNews Daily Bot"
	return messageText
}


func SendMessage(userID int64, username string, articles []Article) (string, error) {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        return "", fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
    }

    apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

    messageText := formatMessage(username, articles)
    data := url.Values{
        "chat_id": {strconv.FormatInt(userID, 10)},
        "text":    {messageText},
		"parse_mode": {"Markdown"},

    }

    resp, err := http.PostForm(apiURL, data)
    if err != nil {
        return "", fmt.Errorf("error sending message: %v", err)
    }
    defer resp.Body.Close()

    return fmt.Sprintf("Message sent to %d successfully", userID), nil
}

// SendToAllUsers sends articles to all users
func SendToAllUsers(ctx context.Context, articles []Article) error {
	users, err := controller.GetUsers()
	if err != nil {
		return fmt.Errorf("error fetching users: %w", err)
	}
	log.Printf("Fetched %d users", len(users))

	var wg sync.WaitGroup
	errChan := make(chan error, len(users))

	for _, user := range users {
		wg.Add(1)
		go func(user controller.User) {
			defer wg.Done()

			userID, err := strconv.ParseInt(user.UserID, 10, 64)
			if err != nil {
				errChan <- fmt.Errorf("error converting userID for %s: %w", user.Username, err)
				return
			}

			if _, err := SendMessage(userID, user.Username, articles); err != nil {
				errChan <- fmt.Errorf("error sending message to %s: %w", user.Username, err)
				return
			}

			log.Printf("Message sent successfully to %s", user.Username)
		}(user)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while sending messages: %v", errors)
	}

	return nil
}

func main() {
	timeStart := time.Now()
	database.InitDB()
	defer database.DB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	articles, err := getArticles("https://news.ycombinator.com/ask")
	if err != nil {
		log.Fatalf("Error fetching articles: %v", err)
	}
	log.Printf("Fetched %d articles", len(articles))

	if err := SendToAllUsers(ctx, articles); err != nil {
		log.Fatalf("Error sending messages: %v", err)
	}

	log.Println("Messages sent successfully to all users")
	log.Printf("Execution time: %v", time.Since(timeStart))

}
