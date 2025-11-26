package main

import (
	"database/sql"
	"log"
	"strconv"
	"time"
)

// StartEmailFetchScheduler starts a background scheduler to fetch emails periodically
func StartEmailFetchScheduler(db *sql.DB) {
	// Get interval from environment (in minutes, default 10 minutes)
	intervalStr := getEnv("EMAIL_FETCH_INTERVAL", "10")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 10
	}

	log.Printf("Starting email fetch scheduler with %d minute interval", interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	go func() {
		for range ticker.C {
			fetchAllProjectEmails(db)
		}
	}()
}

// fetchAllProjectEmails fetches emails for all active projects
func fetchAllProjectEmails(db *sql.DB) {
	log.Println("Starting scheduled email fetch for all active projects")

	rows, err := db.Query("SELECT id, name FROM projects WHERE status = 'active'")
	if err != nil {
		log.Printf("Failed to get active projects: %v", err)
		return
	}
	defer rows.Close()

	processedCount := 0
	failedCount := 0

	for rows.Next() {
		var projectID int
		var projectName string
		if err := rows.Scan(&projectID, &projectName); err != nil {
			log.Printf("Failed to scan project: %v", err)
			continue
		}

		log.Printf("Fetching emails for project %d (%s)", projectID, projectName)
		if err := ProcessEmailsForProject(projectID, db); err != nil {
			log.Printf("Failed to process emails for project %d: %v", projectID, err)
			failedCount++
		} else {
			processedCount++
		}
	}

	log.Printf("Scheduled email fetch completed: %d projects processed, %d failed", processedCount, failedCount)
}
