package controller

import (
	"time"
	"github.com/dapoadedire/hackernews-daily-bot/database"
	_ "github.com/lib/pq"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	UserID    string    `json:"userid"`
	CreatedAt time.Time `json:"created_at"`
}

func GetUsers() ([]User, error) {
	rows, err := database.DB.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Username, &user.UserID, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
