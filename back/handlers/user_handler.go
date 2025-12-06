package handlers

import (
	"db_intro_backend/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UpdateEmailConfig(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		SMTPHost     string `json:"smtp_host"`
		SMTPPort     string `json:"smtp_port"`
		SMTPUsername string `json:"smtp_username"`
		SMTPPassword string `json:"smtp_password"`
		IMAPHost     string `json:"imap_host"`
		IMAPPort     string `json:"imap_port"`
		IMAPUsername string `json:"imap_username"`
		IMAPPassword string `json:"imap_password"`
		EmailAddress string `json:"email_address"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec(`
		UPDATE users 
		SET smtp_host=?, smtp_port=?, smtp_username=?, smtp_password=?, 
			imap_host=?, imap_port=?, imap_username=?, imap_password=?, email_address=?
		WHERE id=?`,
		input.SMTPHost, input.SMTPPort, input.SMTPUsername, input.SMTPPassword,
		input.IMAPHost, input.IMAPPort, input.IMAPUsername, input.IMAPPassword, input.EmailAddress,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email configuration updated successfully"})
}

func GetEmailConfig(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var (
		smtpHost     string
		smtpPort     string
		smtpUsername string
		imapHost     string
		imapPort     string
		imapUsername string
		emailAddress string
	)

	// Use COALESCE to handle NULLs
	err := db.DB.QueryRow(`
		SELECT 
			COALESCE(smtp_host, ''), 
			COALESCE(smtp_port, ''), 
			COALESCE(smtp_username, ''), 
			COALESCE(imap_host, ''), 
			COALESCE(imap_port, ''), 
			COALESCE(imap_username, ''), 
			COALESCE(email_address, '') 
		FROM users WHERE id = ?`, userID).Scan(
		&smtpHost, &smtpPort, &smtpUsername,
		&imapHost, &imapPort, &imapUsername, &emailAddress,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch email config"})
		return
	}

	hasConfig := smtpHost != "" && emailAddress != ""

	c.JSON(http.StatusOK, gin.H{
		"smtp_host":     smtpHost,
		"smtp_port":     smtpPort,
		"smtp_username": smtpUsername,
		"imap_host":     imapHost,
		"imap_port":     imapPort,
		"imap_username": imapUsername,
		"email_address": emailAddress,
		"has_config":    hasConfig,
	})
}
