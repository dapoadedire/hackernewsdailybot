# HackerNews Daily Bot

HackerNews Daily Bot is a Go application that scrapes the latest articles from Hacker News and sends curated content to registered users on Telegram. The bot fetches articles from categories like "Ask HN," "Show HN," and "Top News," delivering personalized messages every morning.

## Features

- Scrapes articles from Hacker News categories.
- Formats articles into a friendly message for Telegram.
- Sends daily updates to multiple Telegram users.
- Optimized with concurrency and error handling.


## Requirements

- Go 1.20+
- PostgreSQL database
- Telegram Bot Token

## Setup Instructions

1. **Clone the Repository**

   ```bash
   git clone https://github.com/dapoadedire/hackernews-daily-bot.git
   cd hackernews-daily-bot

   ```

2. Install Dependencies
   Ensure all required Go modules are installed:

```bash
go mod tidy
```

3. Configure Environment Variables
   Create a .env file in the root directory with the following content:

```bash
TELEGRAM_BOT_TOKEN=<your-telegram-bot-token>
DB_NAME=<your-database-name>
DB_USER=<your-database-username>
DB_PASSWORD=<your-database-password>
DB_HOST=<your-database-host>
DB_PORT=<your-database-port>
```

4.	Set Up the Database

Initialize the PostgreSQL database as required by the database package. 5. Run the Bot
Start the bot:

```bash

go run main.go

```

## Usage

The bot automatically:
- Scrapes articles from Hacker News.
- Formats and sends personalized messages to all registered Telegram users.

## Key Packages Used
-  Colly: Web scraping.
- Godotenv: Environment variable management.


![4F442816-4107-4E41-A1AE-2E610D9C0232](https://github.com/user-attachments/assets/c2a92053-a50c-4aea-8b6c-a026d384c090)
