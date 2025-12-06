package services

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
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

	"db_intro_backend/config"
	"db_intro_backend/db"
	"db_intro_backend/models"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	imapid "github.com/emersion/go-imap-id"
	imapmail "github.com/emersion/go-message/mail"
)

type EmailService struct {
	Config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{Config: cfg}
}

// SendEmail sends an email with optional attachment
func (s *EmailService) SendEmail(to, subject, body string, attachmentPath string) (string, error) {
	from := s.Config.SenderEmail

	// Generate Message-ID
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	domain := "db-intro-system"
	if parts := strings.Split(from, "@"); len(parts) > 1 {
		domain = parts[1]
	}
	messageID := fmt.Sprintf("<%d.%s@%s>", time.Now().Unix(), hex.EncodeToString(randomBytes), domain)

	var msg bytes.Buffer
	writer := multipart.NewWriter(&msg)

	// Email headers
	fmt.Fprintf(&msg, "From: %s\r\n", from)
    fmt.Fprintf(&msg, "To: %s\r\n", to)
    fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
    fmt.Fprintf(&msg, "Message-ID: %s\r\n", messageID)
    fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
    fmt.Fprintf(&msg, "Content-Type: multipart/mixed; boundary=%s\r\n", writer.Boundary())
    fmt.Fprintf(&msg, "\r\n") // header结束

	// Write body part
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/plain; charset=utf-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create body part: %w", err)
	}
	part.Write([]byte(body))

	// Attach file if provided
	if attachmentPath != "" && attachmentPath != "." {
		if err := s.attachFile(writer, attachmentPath); err != nil {
			log.Printf("Warning: Failed to attach file %s: %v", attachmentPath, err)
		}
	}

	writer.Close()

	// Send email
	addr := s.Config.SMTPHost + ":" + s.Config.SMTPPort
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // 可设 false，如果你有证书
		ServerName:         s.Config.SMTPHost,
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return "", fmt.Errorf("TLS dial error: %w", err)
	}
	c, err := smtp.NewClient(conn, s.Config.SMTPHost)
	if err != nil {
		return "", err
	}
	defer c.Quit()
	auth := smtp.PlainAuth("", s.Config.SenderEmail, s.Config.SenderPass, s.Config.SMTPHost)
	if err = c.Auth(auth); err != nil {
		return "", fmt.Errorf("auth error: %w", err)
	}
	c.Mail(from)
	c.Rcpt(to)
	w, err := c.Data()
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return "", err
	}
	_, err = w.Write(msg.Bytes())
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return "", err
	}
	w.Close()

	log.Printf("Email sent successfully to %s", to)
	return messageID, nil
}

func (s *EmailService) attachFile(writer *multipart.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open attachment: %w", err)
	}
	defer file.Close()

	filename := filepath.Base(filePath)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {s.getContentType(filename)},
		"Content-Transfer-Encoding": {"base64"},
		"Content-Disposition":       {fmt.Sprintf(`attachment; filename="%s"`, filename)},
	})
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read attachment: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(content)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		part.Write([]byte(encoded[i:end] + "\r\n"))
	}

	return nil
}

func (s *EmailService) getContentType(filename string) string {
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

// FetchEmails fetches emails from IMAP server
func (s *EmailService) FetchEmails() ([]models.EmailMessage, error) {
	log.Printf("Connecting to IMAP server %s:%s", s.Config.IMAPHost, s.Config.IMAPPort)
	c, err := client.DialTLS(s.Config.IMAPHost+":"+s.Config.IMAPPort, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}
	defer c.Logout()

	if err := c.Login(s.Config.SenderEmail, s.Config.SenderPass); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}
	log.Println("Logged in to IMAP server")

	idClient := imapid.NewClient(c)
    idParams := map[string]string{
        "name":          "my-go-client",
        "version":       "1.0.0",
        "vendor":        "go-imap",
        "support-email": "support@example.com",
    }
	log.Println("Sending IMAP ID...")
    _, err = idClient.ID(idParams)
    if err != nil {
        log.Fatal("ID command failed: ", err)
    }
    log.Println("IMAP ID sent successfully!")

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	if mbox.Messages == 0 {
		log.Println("No messages in mailbox")
		return []models.EmailMessage{}, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchUid}

	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	var emails []models.EmailMessage
	for msg := range messages {
		if msg == nil {
			continue
		}

		r := msg.GetBody(section)
		if r == nil {
			log.Println("Server didn't return message body")
			continue
		}

		emailMsg, err := s.parseEmail(r, msg.Envelope)
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

func (s *EmailService) parseEmail(r io.Reader, envelope *imap.Envelope) (models.EmailMessage, error) {
	mr, err := imapmail.CreateReader(r)
	if err != nil {
		return models.EmailMessage{}, fmt.Errorf("failed to create mail reader: %w", err)
	}

	var emailMsg models.EmailMessage
	emailMsg.RawHeaders = make(map[string]string)

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

	fields := header.Fields()
	for fields.Next() {
		emailMsg.RawHeaders[fields.Key()] = fields.Value()
	}

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
			body, _ := io.ReadAll(p.Body)
			emailMsg.Body += string(body)

		case *imapmail.AttachmentHeader:
			filename, _ := h.Filename()
			contentType, _, _ := h.ContentType()

			data, err := io.ReadAll(p.Body)
			if err != nil {
				log.Printf("Failed to read attachment: %v", err)
				continue
			}

			emailMsg.Attachments = append(emailMsg.Attachments, models.AttachmentInfo{
				Filename:    filename,
				ContentType: contentType,
				Data:        data,
			})
		}
	}

	return emailMsg, nil
}

// ProcessIncomingEmails fetches and processes all new emails
func (s *EmailService) ProcessIncomingEmails() error {
	log.Println("Processing incoming emails...")

	emails, err := s.FetchEmails()
	if err != nil {
		return fmt.Errorf("failed to fetch emails: %w", err)
	}

	if len(emails) == 0 {
		log.Println("No new emails to process")
		return nil
	}

	uploadDir := "./uploads/replies"
	os.MkdirAll(uploadDir, 0755)

	processedCount := 0
	for _, email := range emails {
		var existingID int
		err := db.DB.QueryRow("SELECT id FROM replies WHERE message_id = ?", email.MessageID).Scan(&existingID)
		if err == nil {
			continue
		}

		var projectID int
		var teacherID sql.NullInt64

		if email.InReplyTo != "" {
			var tid int
			log.Printf("Looking up project from In-Reply-To: %s", email.InReplyTo)
			err := db.DB.QueryRow("SELECT project_id, teacher_id FROM sent_emails WHERE message_id = ?", email.InReplyTo).Scan(&projectID, &tid)
			if err == nil {
				teacherID.Int64 = int64(tid)
				teacherID.Valid = true
				log.Printf("Identified project %d from In-Reply-To %s", projectID, email.InReplyTo)
			}
		}

		if projectID == 0 {
			var tid int
			err := db.DB.QueryRow("SELECT id FROM teachers WHERE email = ?", email.From).Scan(&tid)
			if err == nil {
				teacherID.Int64 = int64(tid)
				teacherID.Valid = true

				rows, err := db.DB.Query(`
					SELECT p.id FROM projects p
					JOIN project_members pm ON p.id = pm.project_id
					WHERE pm.teacher_id = ? AND p.status = 'active'
				`, tid)
				if err == nil {
					var pids []int
					for rows.Next() {
						var pid int
						rows.Scan(&pid)
						pids = append(pids, pid)
					}
					rows.Close()

					if len(pids) == 1 {
						projectID = pids[0]
						log.Printf("Identified project %d from sender %s (single active project)", projectID, email.From)
					} else if len(pids) > 1 {
						log.Printf("Ambiguous project for sender %s (multiple active projects: %v)", email.From, pids)
					}
				}
			}
		}

		if projectID == 0 {
			log.Printf("Skipping email %s: could not identify project", email.MessageID)
			continue
		}

		headersJSON := "{}"

		result, err := db.DB.Exec(`
			INSERT INTO replies (project_id, teacher_id, from_email, subject, message_id, in_reply_to, received_at, raw_headers, raw_body)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID, teacherID, email.From, email.Subject, email.MessageID, email.InReplyTo, email.ReceivedAt, headersJSON, email.Body)

		if err != nil {
			log.Printf("Failed to insert reply: %v", err)
			continue
		}

		replyID, _ := result.LastInsertId()

		for _, att := range email.Attachments {
			timestamp := time.Now().Unix()
			safeName := s.sanitizeAttachmentName(att.Filename)
			storedFilename := fmt.Sprintf("%d_%d_%s", projectID, timestamp, safeName)
			storedPath := filepath.Join(uploadDir, storedFilename)

			if err := os.WriteFile(storedPath, att.Data, 0644); err != nil {
				log.Printf("Failed to save attachment %s: %v", att.Filename, err)
				continue
			}

			log.Printf("Saved attachment: %s", storedPath)

			_, err := db.DB.Exec(`
				INSERT INTO attachments (reply_id, project_id, teacher_id, original_filename, stored_path, content_type, file_size)
				VALUES (?, ?, ?, ?, ?, ?, ?)`,
				replyID, projectID, teacherID, att.Filename, storedPath, att.ContentType, len(att.Data))

			if err != nil {
				log.Printf("Failed to insert attachment record: %v", err)
			}
		}

		if teacherID.Valid {
			_, err = db.DB.Exec(`
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

	log.Printf("Processed %d new emails", processedCount)
	return nil
}

func (s *EmailService) sanitizeAttachmentName(name string) string {
	cleaned := filepath.Base(strings.TrimSpace(name))
	cleaned = strings.ReplaceAll(cleaned, "..", "")
	if cleaned == "" {
		return "attachment"
	}

	var b strings.Builder
	for i := 0; i < len(cleaned); i++ {
		ch := cleaned[i]
		isLetter := (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
		isDigit := ch >= '0' && ch <= '9'
		switch {
		case isLetter || isDigit:
			b.WriteByte(ch)
		case ch == '.' || ch == '-' || ch == '_':
			b.WriteByte(ch)
		default:
			b.WriteByte('_')
		}
	}

	sanitized := strings.Trim(b.String(), "_")
	if sanitized == "" {
		sanitized = "attachment"
	}
	return sanitized
}
