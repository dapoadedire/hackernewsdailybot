package database

import (
	"database/sql"
	"fmt"

	"github.com/dapoadedire/hackernews-daily-bot/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	config.LoadEnv()

	DB_NAME, DB_USER, DB_PASSWORD, DB_HOST, DB_PORT := config.GetDBConfig()

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", DB_HOST, DB_PORT, DB_USER, DB_NAME, DB_PASSWORD)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	if err = DB.Ping(); err != nil {
		panic(err)
	}
	_, err = DB.Exec(`
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; 
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), 
        username TEXT, 
        userid TEXT, 
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
`)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to the database!")
}
