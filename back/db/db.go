package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"db_intro_backend/config"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB(cfg *config.Config) {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&charset=utf8mb4",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBName,
	)

	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Retry connection with exponential backoff
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		err = DB.Ping()
		if err == nil {
			log.Println("Database connected successfully")
			return
		}

		if i == maxRetries-1 {
			log.Fatal("Failed to connect to database after retries:", err)
		}

		waitTime := time.Duration(i+1) * time.Second
		log.Printf("Database not ready, retrying in %v... (attempt %d/%d)", waitTime, i+1, maxRetries)
		time.Sleep(waitTime)
	}
}
