package config

import (
	"db_intro_backend/utils"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBName     string
	Port       string
	
	SMTPHost    string
	SMTPPort    string
	SenderEmail string
	SenderPass  string
	IMAPHost    string
	IMAPPort    string
	
	EmailFetchInterval int
}

func LoadConfig() *Config {
	return &Config{
		DBUser:     utils.GetEnv("DB_USER", "root"),
		DBPassword: utils.GetEnv("DB_PASSWORD", "root"),
		DBHost:     utils.GetEnv("DB_HOST", "localhost"),
		DBName:     utils.GetEnv("DB_NAME", "db_intro"),
		Port:       utils.GetEnv("PORT", "8080"),
		
		SMTPHost:    utils.GetEnv("SMTP_HOST", "smtp.example.com"),
		SMTPPort:    utils.GetEnv("SMTP_PORT", "587"),
		SenderEmail: utils.GetEnv("SENDER_EMAIL", "noreply@example.com"),
		SenderPass:  utils.GetEnv("SENDER_PASS", ""),
		IMAPHost:    utils.GetEnv("IMAP_HOST", "imap.example.com"),
		IMAPPort:    utils.GetEnv("IMAP_PORT", "993"),
	}
}
