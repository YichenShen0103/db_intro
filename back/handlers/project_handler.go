package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"db_intro_backend/db"
	"db_intro_backend/models"
	"db_intro_backend/services"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	EmailService *services.EmailService
	ExcelService *services.ExcelService
}

func NewProjectHandler(emailService *services.EmailService, excelService *services.ExcelService) *ProjectHandler {
	return &ProjectHandler{
		EmailService: emailService,
		ExcelService: excelService,
	}
}

func (h *ProjectHandler) GetProjects(c *gin.Context) {
	userID := c.GetInt("userID")
	query := `
		SELECT 
			p.id, p.code, p.name, p.status, p.email_subject_template, 
			p.email_body_template, p.excel_template_filename, p.created_at,
			COUNT(DISTINCT pm.id) as total_sent,
			COUNT(DISTINCT CASE WHEN pm.current_status = 'replied' THEN pm.id END) as replied_count
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.created_by = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Status, &p.EmailSubjectTemplate,
			&p.EmailBodyTemplate, &p.ExcelTemplateFilename, &p.CreatedAt,
			&p.TotalSent, &p.RepliedCount); err != nil {
			continue
		}
		projects = append(projects, p)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": projects})
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID := c.GetInt("userID")
	id := c.Param("id")
	var p models.Project
	err := db.DB.QueryRow(`
		SELECT 
			p.id, p.code, p.name, p.status, p.email_subject_template,
			p.email_body_template, p.excel_template_filename, p.created_at,
			COALESCE(stats.total_sent, 0) AS total_sent,
			COALESCE(stats.replied_count, 0) AS replied_count
		FROM projects p
		LEFT JOIN (
			SELECT project_id,
				COUNT(*) AS total_sent,
				COUNT(CASE WHEN current_status = 'replied' THEN 1 END) AS replied_count
			FROM project_members
			GROUP BY project_id
		) stats ON stats.project_id = p.id
		WHERE p.id=? AND p.created_by=?
	`, id, userID).Scan(&p.ID, &p.Code, &p.Name, &p.Status, &p.EmailSubjectTemplate,
		&p.EmailBodyTemplate, &p.ExcelTemplateFilename, &p.CreatedAt, &p.TotalSent, &p.RepliedCount)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": p})
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	name := c.PostForm("name")
	code := c.PostForm("code")
	emailSubject := c.PostForm("email_subject_template")
	emailBody := c.PostForm("email_body_template")
	log.Printf("Creating project with name: %s, code: %s", name, code)

	// Handle file upload
	file, err := c.FormFile("excel_template")
	var filename string
	if err == nil {
		// Create uploads directory if not exists
		uploadDir := "./uploads/templates"
		os.MkdirAll(uploadDir, 0755)

		filename = fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		dst := filepath.Join(uploadDir, filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
	} else {
		log.Printf("No file uploaded for project %s: %v", name, err)
	}

	userID := c.GetInt("userID")
	result, err := db.DB.Exec(
		"INSERT INTO projects (code, name, email_subject_template, email_body_template, excel_template_filename, created_by) VALUES (?, ?, ?, ?, ?, ?)",
		code, name, emailSubject, emailBody, filename, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"id": id}})
}

func (h *ProjectHandler) DispatchProject(c *gin.Context) {
	userID := c.GetInt("userID")
	projectID := c.Param("id")

	// Verify ownership
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND created_by = ?", projectID, userID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	// Query project members who haven't been sent an email yet
	rows, err := db.DB.Query("SELECT teacher_id FROM project_members WHERE project_id = ? AND sent_at IS NULL", projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var teacherIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		teacherIDs = append(teacherIDs, id)
	}

	if len(teacherIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "No pending emails to send", "target_count": 0})
		return
	}

	// Get project details for email template
	var project models.Project
	err = db.DB.QueryRow(`
		SELECT id, code, name, email_subject_template, email_body_template, excel_template_filename, created_by
		FROM projects WHERE id=?
	`, projectID).Scan(&project.ID, &project.Code, &project.Name, &project.EmailSubjectTemplate,
		&project.EmailBodyTemplate, &project.ExcelTemplateFilename, &project.CreatedBy)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project details"})
		return
	}

	// Prepare attachment path
	var attachmentPath string
	if project.ExcelTemplateFilename != "" {
		attachmentPath = filepath.Join("./uploads/templates", project.ExcelTemplateFilename)
	}

	targetType := "pending_members"
	h.dispatchProjectEmailsAsync(project, teacherIDs, attachmentPath, targetType)

	c.JSON(http.StatusAccepted, gin.H{
		"code":         202,
		"message":      "Email dispatch started",
		"target_count": len(teacherIDs),
	})
}

func (h *ProjectHandler) GetProjectTracking(c *gin.Context) {
	userID := c.GetInt("userID")
	projectID := c.Param("id")

	// Verify ownership
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND created_by = ?", projectID, userID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	query := `
		SELECT 
			t.id, t.name, COALESCE(d.name, '未指定'), 
			pm.current_status, pm.last_reply_at
		FROM project_members pm
		JOIN teachers t ON pm.teacher_id = t.id
		LEFT JOIN departments d ON t.department_id = d.id
		WHERE pm.project_id = ?
	`

	rows, err := db.DB.Query(query, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var records []models.TrackingRecord
	totalSent := 0
	repliedCount := 0

	for rows.Next() {
		var r models.TrackingRecord
		var replyTime sql.NullTime
		if err := rows.Scan(&r.TeacherID, &r.Name, &r.Department, &r.Status, &replyTime); err != nil {
			continue
		}
		if replyTime.Valid {
			timeStr := replyTime.Time.Format("2006-01-02 15:04:05")
			r.ReplyTime = &timeStr
		}
		records = append(records, r)
		totalSent++
		if r.Status == "replied" {
			repliedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total_sent":    totalSent,
			"replied_count": repliedCount,
			"details":       records,
		},
	})
}

func (h *ProjectHandler) RemindTeachers(c *gin.Context) {
	userID := c.GetInt("userID")
	projectID := c.Param("id")

	// Verify ownership
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND created_by = ?", projectID, userID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	var req struct {
		TargetIDs []int `json:"target_ids,omitempty"`
	}
	c.ShouldBindJSON(&req)

	query := "SELECT t.id, t.name, t.email FROM project_members pm JOIN teachers t ON pm.teacher_id = t.id WHERE pm.project_id = ? AND pm.current_status = 'pending'"

	if len(req.TargetIDs) > 0 {
		query += " AND t.id IN ("
		for i := range req.TargetIDs {
			if i > 0 {
				query += ","
			}
			query += "?"
		}
		query += ")"
	}

	var args []interface{}
	args = append(args, projectID)
	for _, id := range req.TargetIDs {
		args = append(args, id)
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var targets []reminderTarget
	for rows.Next() {
		var t reminderTarget
		if err := rows.Scan(&t.ID, &t.Name, &t.Email); err != nil {
			log.Printf("Failed to scan reminder target: %v", err)
			continue
		}
		targets = append(targets, t)
	}

	if len(targets) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "No pending teachers to remind", "count": 0})
		return
	}

	// Get project details for reminder template
	var project models.Project
	err = db.DB.QueryRow(`
		SELECT id, code, name, email_subject_template, email_body_template, created_by
		FROM projects WHERE id=?
	`, projectID).Scan(&project.ID, &project.Code, &project.Name,
		&project.EmailSubjectTemplate, &project.EmailBodyTemplate, &project.CreatedBy)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project details"})
		return
	}

	h.sendRemindersAsync(project, targets)

	c.JSON(http.StatusAccepted, gin.H{
		"code":         202,
		"message":      "Reminder emails queued",
		"target_count": len(targets),
	})
}

func (h *ProjectHandler) dispatchProjectEmailsAsync(project models.Project, teacherIDs []int, attachmentPath, targetType string) {
	ids := append([]int(nil), teacherIDs...)
	go func(p models.Project, ids []int, attach, tType string) {
		// Fetch user config
		var user models.User
		var smtpHost, smtpPort, smtpUser, smtpPass, imapHost, imapPort, imapUser, imapPass, emailAddr sql.NullString
		err := db.DB.QueryRow("SELECT id, smtp_host, smtp_port, smtp_username, smtp_password, imap_host, imap_port, imap_username, imap_password, email_address FROM users WHERE id = ?", p.CreatedBy).Scan(
			&user.ID, &smtpHost, &smtpPort, &smtpUser, &smtpPass, &imapHost, &imapPort, &imapUser, &imapPass, &emailAddr,
		)
		if err != nil {
			log.Printf("Failed to fetch user config for project %d: %v", p.ID, err)
			return
		}
		user.SMTPHost = smtpHost.String
		user.SMTPPort = smtpPort.String
		user.SMTPUsername = smtpUser.String
		user.SMTPPassword = smtpPass.String
		user.IMAPHost = imapHost.String
		user.IMAPPort = imapPort.String
		user.IMAPUsername = imapUser.String
		user.IMAPPassword = imapPass.String
		user.EmailAddress = emailAddr.String

		if user.SMTPHost == "" || user.EmailAddress == "" {
			log.Printf("User %d has no email config, aborting dispatch", p.CreatedBy)
			return
		}

		log.Printf("Starting background dispatch for project %d (%d teachers)...", p.ID, len(ids))
		successCount := 0
		for _, tid := range ids {
			var teacherEmail, teacherName string
			if err := db.DB.QueryRow("SELECT name, email FROM teachers WHERE id=?", tid).Scan(&teacherName, &teacherEmail); err != nil {
				log.Printf("Failed to get teacher %d: %v", tid, err)
				continue
			}

			subject := h.replaceTemplateVars(p.EmailSubjectTemplate, teacherName, p.Name)
			body := h.replaceTemplateVars(p.EmailBodyTemplate, teacherName, p.Name)

			msgID, err := h.EmailService.SendEmail(user, teacherEmail, subject, body, attach)
			if err != nil {
				log.Printf("Failed to send email to %s: %v", teacherEmail, err)
				continue
			}

			if _, err := db.DB.Exec("INSERT INTO sent_emails (project_id, teacher_id, message_id) VALUES (?, ?, ?)", p.ID, tid, msgID); err != nil {
				log.Printf("Failed to record sent email for teacher %d: %v", tid, err)
			}

			if _, err := db.DB.Exec(
				"INSERT INTO project_members (project_id, teacher_id, sent_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE sent_at=?",
				p.ID, tid, time.Now(), time.Now(),
			); err != nil {
				log.Printf("Failed to upsert project member %d: %v", tid, err)
				continue
			}

			successCount++
		}

		if _, err := db.DB.Exec(
			"INSERT INTO dispatches (project_id, target_type, sent_count) VALUES (?, ?, ?)",
			p.ID, tType, successCount,
		); err != nil {
			log.Printf("Failed to record dispatch for project %d: %v", p.ID, err)
		}

		log.Printf("Background dispatch for project %d finished: %d/%d succeeded", p.ID, successCount, len(ids))
	}(project, ids, attachmentPath, targetType)
}

type reminderTarget struct {
	ID    int
	Name  string
	Email string
}

func (h *ProjectHandler) sendRemindersAsync(project models.Project, targets []reminderTarget) {
	targetCopies := append([]reminderTarget(nil), targets...)
	go func(p models.Project, items []reminderTarget) {
		// Fetch user config
		var user models.User
		var smtpHost, smtpPort, smtpUser, smtpPass, imapHost, imapPort, imapUser, imapPass, emailAddr sql.NullString
		err := db.DB.QueryRow("SELECT id, smtp_host, smtp_port, smtp_username, smtp_password, imap_host, imap_port, imap_username, imap_password, email_address FROM users WHERE id = ?", p.CreatedBy).Scan(
			&user.ID, &smtpHost, &smtpPort, &smtpUser, &smtpPass, &imapHost, &imapPort, &imapUser, &imapPass, &emailAddr,
		)
		if err != nil {
			log.Printf("Failed to fetch user config for project %d: %v", p.ID, err)
			return
		}
		user.SMTPHost = smtpHost.String
		user.SMTPPort = smtpPort.String
		user.SMTPUsername = smtpUser.String
		user.SMTPPassword = smtpPass.String
		user.IMAPHost = imapHost.String
		user.IMAPPort = imapPort.String
		user.IMAPUsername = imapUser.String
		user.IMAPPassword = imapPass.String
		user.EmailAddress = emailAddr.String

		if user.SMTPHost == "" || user.EmailAddress == "" {
			log.Printf("User %d has no email config, aborting reminders", p.CreatedBy)
			return
		}

		log.Printf("Starting reminder run for project %d (%d targets)...", p.ID, len(items))
		successCount := 0
		for _, t := range items {
			subject := "催促提醒: " + p.EmailSubjectTemplate
			body := fmt.Sprintf("尊敬的%s老师：\n\n这是一封催促提醒邮件。\n\n%s\n\n请尽快完成并回复，谢谢！\n\n原邮件内容：\n%s",
				t.Name, p.Name, p.EmailBodyTemplate)

			msgID, err := h.EmailService.SendEmail(user, t.Email, subject, body, "")
			if err != nil {
				log.Printf("Failed to send reminder to %s (%s): %v", t.Name, t.Email, err)
				continue
			}

			if _, err := db.DB.Exec("INSERT INTO sent_emails (project_id, teacher_id, message_id) VALUES (?, ?, ?)", p.ID, t.ID, msgID); err != nil {
				log.Printf("Failed to record reminder for teacher %d: %v", t.ID, err)
				continue
			}

			log.Printf("Reminder sent to %s (%s)", t.Name, t.Email)
			successCount++
		}

		log.Printf("Reminder run for project %d finished: %d/%d succeeded", p.ID, successCount, len(items))
	}(project, targetCopies)
}

func (h *ProjectHandler) FetchProjectEmails(c *gin.Context) {
	userID := c.GetInt("userID")
	projectID := c.Param("id")

	// Convert projectID to int
	var pid int
	if _, err := fmt.Sscanf(projectID, "%d", &pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify ownership
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND created_by = ?", pid, userID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	// Process emails only for the current user
	if err := h.EmailService.ProcessUserEmails(userID); err != nil {
		log.Printf("Failed to process emails for user %d project %d: %v", userID, pid, err)
		status := http.StatusInternalServerError
		message := "Failed to fetch emails"
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			status = http.StatusNotFound
			message = "User not found"
		case errors.Is(err, services.ErrEmailConfigIncomplete):
			status = http.StatusBadRequest
			message = "Email configuration incomplete"
		}
		c.JSON(status, gin.H{
			"code":    status,
			"error":   message,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Emails fetched and processed successfully",
	})
}

func (h *ProjectHandler) AggregateData(c *gin.Context) {
	userID := c.GetInt("userID")
	projectID := c.Param("id")
	pid, err := strconv.Atoi(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify ownership
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND created_by = ?", pid, userID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	outputPath, attachmentCount, rowCount, err := h.ExcelService.AggregateProjectExcel(pid)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, services.ErrNoExcelAttachments) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Aggregated %d rows from %d attachments for project %d", rowCount, attachmentCount, pid)
	c.JSON(http.StatusOK, gin.H{
		"code":        200,
		"message":     "Aggregation completed",
		"attachments": attachmentCount,
		"rows":        rowCount,
		"file_path":   outputPath,
	})
}

func (h *ProjectHandler) DownloadAggregated(c *gin.Context) {
	userID := c.GetInt("userID")
	projectID := c.Param("id")

	// Verify ownership
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND created_by = ?", projectID, userID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	filePath := h.ExcelService.AggregatedFilePath(projectID)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregated file not found"})
		return
	}

	c.FileAttachment(filePath, fmt.Sprintf("project_%s_aggregated.xlsx", projectID))
}

// replaceTemplateVars replaces template variables in email content
func (h *ProjectHandler) replaceTemplateVars(text, teacherName, projectName string) string {
	// Support common template variables
	text = strings.ReplaceAll(text, "{{teacher_name}}", teacherName)
	text = strings.ReplaceAll(text, "{{project_name}}", projectName)
	text = strings.ReplaceAll(text, "{{Teacher_Name}}", teacherName)
	text = strings.ReplaceAll(text, "{{Project_Name}}", projectName)
	return text
}

func (h *ProjectHandler) AddProjectMembers(c *gin.Context) {
	userID := c.GetInt("userID")
	projectID := c.Param("id")

	// Verify ownership
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ? AND created_by = ?", projectID, userID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	var req struct {
		TeacherIDs []int `json:"teacher_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.TeacherIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No teacher IDs provided"})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT IGNORE INTO project_members (project_id, teacher_id) VALUES (?, ?)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	addedCount := 0
	for _, tid := range req.TeacherIDs {
		res, err := stmt.Exec(projectID, tid)
		if err != nil {
			log.Printf("Failed to add member %d to project %s: %v", tid, projectID, err)
			continue
		}
		affected, _ := res.RowsAffected()
		addedCount += int(affected)
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":        200,
		"message":     "Members added successfully",
		"added_count": addedCount,
	})
}
