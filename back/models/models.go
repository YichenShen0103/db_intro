package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Don't return password in JSON
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID                    int       `json:"id"`
	Code                  string    `json:"code"`
	Name                  string    `json:"name"`
	Status                string    `json:"status"`
	EmailSubjectTemplate  string    `json:"email_subject_template"`
	EmailBodyTemplate     string    `json:"email_body_template"`
	ExcelTemplateFilename string    `json:"excel_template_filename"`
	CreatedBy             int       `json:"created_by"`
	CreatedAt             time.Time `json:"created_at"`
	TotalSent             int       `json:"total_sent"`
	RepliedCount          int       `json:"replied_count"`
}

type Teacher struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	DepartmentID   *int      `json:"department_id"`
	DepartmentName string    `json:"department_name,omitempty"`
	Phone          string    `json:"phone"`
	CreatedAt      time.Time `json:"created_at"`
}

type Department struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type TrackingRecord struct {
	TeacherID  int     `json:"teacher_id"`
	Name       string  `json:"name"`
	Department string  `json:"department"`
	Status     string  `json:"status"`
	ReplyTime  *string `json:"reply_time"`
}

type AttachmentMeta struct {
	StoredPath     string
	OriginalName   string
	TeacherName    string
	TeacherEmail   string
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
