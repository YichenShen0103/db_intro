package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	imapmail "github.com/emersion/go-message/mail"
)

// EmailConfig holds email server configuration
type EmailConfig struct {
	SMTPHost    string
	SMTPPort    string
	SenderEmail string
	SenderPass  string
	IMAPHost    string
	IMAPPort    string
}

// GetEmailConfig reads email configuration from environment
func GetEmailConfig() EmailConfig {
	return EmailConfig{
		SMTPHost:    getEnv("SMTP_HOST", "smtp.example.com"),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SenderEmail: getEnv("SENDER_EMAIL", "noreply@example.com"),
		SenderPass:  getEnv("SENDER_PASS", ""),
		IMAPHost:    getEnv("IMAP_HOST", "imap.example.com"),
		IMAPPort:    getEnv("IMAP_PORT", "993"),
	}
}

// SendEmail sends an email with optional attachment
func SendEmail(to, subject, body string, attachmentPath string) error {
	config := GetEmailConfig()
	from := config.SenderEmail

	var msg bytes.Buffer

	// Email headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	// Create multipart writer
	writer := multipart.NewWriter(&msg)
	boundary := writer.Boundary()
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	msg.WriteString("\r\n")

	// Write body part
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/plain; charset=utf-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return fmt.Errorf("failed to create body part: %w", err)
	}
	part.Write([]byte(body))

	// Attach file if provided
	if attachmentPath != "" && attachmentPath != "." {
		if err := attachFile(writer, attachmentPath); err != nil {
			log.Printf("Warning: Failed to attach file %s: %v", attachmentPath, err)
		}
	}

	writer.Close()

	// Authentication
	auth := smtp.PlainAuth("", config.SenderEmail, config.SenderPass, config.SMTPHost)

	// Send email
	addr := config.SMTPHost + ":" + config.SMTPPort
	err = smtp.SendMail(addr, auth, from, []string{to}, msg.Bytes())
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}

// attachFile attaches a file to the multipart writer
func attachFile(writer *multipart.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open attachment: %w", err)
	}
	defer file.Close()

	filename := filepath.Base(filePath)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {getContentType(filename)},
		"Content-Transfer-Encoding": {"base64"},
		"Content-Disposition":       {fmt.Sprintf(`attachment; filename="%s"`, filename)},
	})
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}

	// Read and encode file
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read attachment: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(content)
	// Split into 76-character lines as per RFC 2045
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		part.Write([]byte(encoded[i:end] + "\r\n"))
	}

	return nil
}

// getContentType returns the MIME content type for a file
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// SendBulkEmails sends emails to multiple recipients
func SendBulkEmails(recipients []string, subject, body string, attachmentPath string) (int, error) {
	successCount := 0
	for _, recipient := range recipients {
		if err := SendEmail(recipient, subject, body, attachmentPath); err != nil {
			log.Printf("Failed to send to %s: %v", recipient, err)
			continue
		}
		successCount++
	}
	return successCount, nil
}

// EmailMessage represents a received email
type EmailMessage struct {
	MessageID   string
	From        string
	Subject     string
	Body        string
	InReplyTo   string
	Attachments []AttachmentInfo
	ReceivedAt  time.Time
	RawHeaders  map[string]string
}

// AttachmentInfo represents an email attachment
type AttachmentInfo struct {
	Filename    string
	ContentType string
	Data        []byte
}

// FetchEmails fetches emails from IMAP server for a specific project
func FetchEmails(projectID int) ([]EmailMessage, error) {
	config := GetEmailConfig()

	// Connect to IMAP server
	log.Printf("Connecting to IMAP server %s:%s", config.IMAPHost, config.IMAPPort)
	c, err := client.DialTLS(config.IMAPHost+":"+config.IMAPPort, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.SenderEmail, config.SenderPass); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}
	log.Println("Logged in to IMAP server")

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	if mbox.Messages == 0 {
		log.Println("No messages in mailbox")
		return []EmailMessage{}, nil
	}

	// Get all unseen messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchUid}

	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	var emails []EmailMessage
	for msg := range messages {
		if msg == nil {
			continue
		}

		r := msg.GetBody(section)
		if r == nil {
			log.Println("Server didn't return message body")
			continue
		}

		// Parse email
		emailMsg, err := parseEmail(r, msg.Envelope)
		if err != nil {
			log.Printf("Failed to parse email: %v", err)
			continue
		}

		emails = append(emails, emailMsg)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	log.Printf("Fetched %d emails", len(emails))
	return emails, nil
}

// parseEmail parses a raw email message
func parseEmail(r io.Reader, envelope *imap.Envelope) (EmailMessage, error) {
	mr, err := imapmail.CreateReader(r)
	if err != nil {
		return EmailMessage{}, fmt.Errorf("failed to create mail reader: %w", err)
	}

	var emailMsg EmailMessage
	emailMsg.RawHeaders = make(map[string]string)

	// Parse headers
	header := mr.Header
	if date, err := header.Date(); err == nil {
		emailMsg.ReceivedAt = date
	} else {
		emailMsg.ReceivedAt = time.Now()
	}

	if from, err := header.AddressList("From"); err == nil && len(from) > 0 {
		emailMsg.From = from[0].Address
	}

	emailMsg.Subject, _ = header.Subject()
	emailMsg.MessageID = header.Get("Message-ID")
	emailMsg.InReplyTo = header.Get("In-Reply-To")

	// Store raw headers
	fields := header.Fields()
	for fields.Next() {
		emailMsg.RawHeaders[fields.Key()] = fields.Value()
	}

	// Parse body and attachments
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Error reading part: %v", err)
			continue
		}

		switch h := p.Header.(type) {
		case *imapmail.InlineHeader:
			// Read body
			body, _ := io.ReadAll(p.Body)
			emailMsg.Body += string(body)

		case *imapmail.AttachmentHeader:
			// Read attachment
			filename, _ := h.Filename()
			contentType, _, _ := h.ContentType()

			data, err := io.ReadAll(p.Body)
			if err != nil {
				log.Printf("Failed to read attachment: %v", err)
				continue
			}

			emailMsg.Attachments = append(emailMsg.Attachments, AttachmentInfo{
				Filename:    filename,
				ContentType: contentType,
				Data:        data,
			})
		}
	}

	return emailMsg, nil
}

// ProcessEmailsForProject fetches and processes emails for a project
func ProcessEmailsForProject(projectID int, db *sql.DB) error {
	log.Printf("Processing emails for project %d", projectID)

	// Fetch emails from IMAP
	emails, err := FetchEmails(projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch emails: %w", err)
	}

	if len(emails) == 0 {
		log.Println("No new emails to process")
		return nil
	}

	// Create upload directory
	uploadDir := "./uploads/replies"
	os.MkdirAll(uploadDir, 0755)

	processedCount := 0
	for _, email := range emails {
		// Check if email already exists
		var existingID int
		err := db.QueryRow("SELECT id FROM replies WHERE message_id = ?", email.MessageID).Scan(&existingID)
		if err == nil {
			log.Printf("Email %s already processed, skipping", email.MessageID)
			continue
		}

		// Find teacher by email
		var teacherID sql.NullInt64
		err = db.QueryRow("SELECT id FROM teachers WHERE email = ?", email.From).Scan(&teacherID)
		if err != nil {
			log.Printf("Teacher not found for email %s", email.From)
		}

		// Convert headers to JSON string (simplified)
		headersJSON := "{}"

		// Insert reply record
		result, err := db.Exec(`
			INSERT INTO replies (project_id, teacher_id, from_email, subject, message_id, in_reply_to, received_at, raw_headers, raw_body)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, teacherID, email.From, email.Subject, email.MessageID, email.InReplyTo, email.ReceivedAt, headersJSON, email.Body)

		if err != nil {
			log.Printf("Failed to insert reply: %v", err)
			continue
		}

		replyID, _ := result.LastInsertId()

		// Process attachments
		for _, att := range email.Attachments {
			// Save attachment to disk
			timestamp := time.Now().Unix()
			storedFilename := fmt.Sprintf("%d_%d_%s", projectID, timestamp, att.Filename)
			storedPath := filepath.Join(uploadDir, storedFilename)

			if err := os.WriteFile(storedPath, att.Data, 0644); err != nil {
				log.Printf("Failed to save attachment %s: %v", att.Filename, err)
				continue
			}

			log.Printf("Saved attachment: %s", storedPath)

			// Determine Excel type
			excelType := "unknown"
			if strings.HasSuffix(strings.ToLower(att.Filename), ".xlsx") ||
				strings.HasSuffix(strings.ToLower(att.Filename), ".xls") {
				excelType = detectExcelType(storedPath)
			}

			// Insert attachment record
			_, err := db.Exec(`
				INSERT INTO attachments (reply_id, project_id, teacher_id, original_filename, stored_path, content_type, file_size, excel_type)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				replyID, projectID, teacherID, att.Filename, storedPath, att.ContentType, len(att.Data), excelType)

			if err != nil {
				log.Printf("Failed to insert attachment record: %v", err)
			}
		}

		// Update project_members status
		if teacherID.Valid {
			_, err = db.Exec(`
				UPDATE project_members 
				SET current_status = 'replied', last_reply_at = ?
				WHERE project_id = ? AND teacher_id = ?`,
				email.ReceivedAt, projectID, teacherID.Int64)

			if err != nil {
				log.Printf("Failed to update project_members: %v", err)
			}
		}

		processedCount++
	}

	log.Printf("Processed %d new emails for project %d", processedCount, projectID)
	return nil
}

// detectExcelType attempts to determine if the Excel file is type_a or type_b
func detectExcelType(filePath string) string {
	// TODO: Implement logic to detect Excel type based on content
	// For now, return unknown
	// You can use a library like excelize to read and analyze the Excel structure
	return "unknown"
}
