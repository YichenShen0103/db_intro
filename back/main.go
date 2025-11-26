package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Project struct {
	ID                    int       `json:"id"`
	Code                  string    `json:"code"`
	Name                  string    `json:"name"`
	Status                string    `json:"status"`
	EmailSubjectTemplate  string    `json:"email_subject_template"`
	EmailBodyTemplate     string    `json:"email_body_template"`
	ExcelTemplateFilename string    `json:"excel_template_filename"`
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

func main() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&charset=utf8mb4",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", "root"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_NAME", "db_front"),
	)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Retry connection with exponential backoff
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Database connected successfully")
			break
		}

		if i == maxRetries-1 {
			log.Fatal("Failed to connect to database after retries:", err)
		}

		waitTime := time.Duration(i+1) * time.Second
		log.Printf("Database not ready, retrying in %v... (attempt %d/%d)", waitTime, i+1, maxRetries)
		time.Sleep(waitTime)
	}

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Static file serving for uploads
	r.Static("/uploads", "./uploads")

	// Start email fetch scheduler
	if getEnv("ENABLE_EMAIL_SCHEDULER", "true") == "true" {
		StartEmailFetchScheduler(db)
	}

	// API routes
	api := r.Group("/api")
	{
		// Health check
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// Departments
		api.GET("/departments", getDepartments)

		// Teachers
		api.GET("/teachers", getTeachers)
		api.POST("/teachers", createTeacher)
		api.PUT("/teachers/:id", updateTeacher)
		api.DELETE("/teachers/:id", deleteTeacher)

		// Projects
		api.GET("/projects", getProjects)
		api.GET("/projects/:id", getProject)
		api.POST("/projects", createProject)
		api.POST("/projects/:id/dispatch", dispatchProject)
		api.GET("/projects/:id/tracking", getProjectTracking)
		api.POST("/projects/:id/remind", remindTeachers)
		api.POST("/projects/:id/fetch-emails", fetchProjectEmails)
		api.POST("/projects/:id/aggregate", aggregateData)
		api.GET("/projects/:id/download", downloadAggregated)
	}

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s...\n", port)
	r.Run(":" + port)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// replaceTemplateVars replaces template variables in email content
func replaceTemplateVars(text, teacherName, projectName string) string {
	// Support common template variables
	text = strings.ReplaceAll(text, "{{teacher_name}}", teacherName)
	text = strings.ReplaceAll(text, "{{project_name}}", projectName)
	text = strings.ReplaceAll(text, "{{Teacher_Name}}", teacherName)
	text = strings.ReplaceAll(text, "{{Project_Name}}", projectName)
	return text
}

// Departments
func getDepartments(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, code, created_at FROM departments")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var departments []Department
	for rows.Next() {
		var d Department
		if err := rows.Scan(&d.ID, &d.Name, &d.Code, &d.CreatedAt); err != nil {
			continue
		}
		departments = append(departments, d)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": departments})
}

// Teachers
func getTeachers(c *gin.Context) {
	query := `
		SELECT t.id, t.name, t.email, t.department_id, d.name as department_name, t.phone, t.created_at 
		FROM teachers t
		LEFT JOIN departments d ON t.department_id = d.id
	`
	departmentFilter := c.Query("department")
	if departmentFilter != "" {
		query += " WHERE d.name = ?"
	}

	var rows *sql.Rows
	var err error
	if departmentFilter != "" {
		rows, err = db.Query(query, departmentFilter)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var teachers []Teacher
	for rows.Next() {
		var t Teacher
		var deptName sql.NullString
		if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.DepartmentID, &deptName, &t.Phone, &t.CreatedAt); err != nil {
			continue
		}
		if deptName.Valid {
			t.DepartmentName = deptName.String
		}
		teachers = append(teachers, t)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": teachers})
}

func createTeacher(c *gin.Context) {
	var t Teacher
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec(
		"INSERT INTO teachers (name, email, department_id, phone) VALUES (?, ?, ?, ?)",
		t.Name, t.Email, t.DepartmentID, t.Phone,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"id": id}})
}

func updateTeacher(c *gin.Context) {
	id := c.Param("id")
	var t Teacher
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(
		"UPDATE teachers SET name=?, email=?, department_id=?, phone=? WHERE id=?",
		t.Name, t.Email, t.DepartmentID, t.Phone, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "updated"})
}

func deleteTeacher(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM teachers WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

// Projects
func getProjects(c *gin.Context) {
	query := `
		SELECT 
			p.id, p.code, p.name, p.status, p.email_subject_template, 
			p.email_body_template, p.excel_template_filename, p.created_at,
			COUNT(DISTINCT pm.id) as total_sent,
			COUNT(DISTINCT CASE WHEN pm.current_status = 'replied' THEN pm.id END) as replied_count
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Status, &p.EmailSubjectTemplate,
			&p.EmailBodyTemplate, &p.ExcelTemplateFilename, &p.CreatedAt,
			&p.TotalSent, &p.RepliedCount); err != nil {
			continue
		}
		projects = append(projects, p)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": projects})
}

func getProject(c *gin.Context) {
	id := c.Param("id")
	var p Project
	err := db.QueryRow(`
		SELECT id, code, name, status, email_subject_template, 
		email_body_template, excel_template_filename, created_at
		FROM projects WHERE id=?
	`, id).Scan(&p.ID, &p.Code, &p.Name, &p.Status, &p.EmailSubjectTemplate,
		&p.EmailBodyTemplate, &p.ExcelTemplateFilename, &p.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": p})
}

func createProject(c *gin.Context) {
	name := c.PostForm("name")
	code := c.PostForm("code")
	emailSubject := c.PostForm("email_subject_template")
	emailBody := c.PostForm("email_body_template")

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
	}

	result, err := db.Exec(
		"INSERT INTO projects (code, name, email_subject_template, email_body_template, excel_template_filename) VALUES (?, ?, ?, ?, ?)",
		code, name, emailSubject, emailBody, filename,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"id": id}})
}

func dispatchProject(c *gin.Context) {
	projectID := c.Param("id")

	var req struct {
		TargetType   string `json:"target_type"` // all, department, selected
		DepartmentID int    `json:"department_id,omitempty"`
		TeacherIDs   []int  `json:"teacher_ids,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get teacher IDs based on target type
	var teacherIDs []int
	switch req.TargetType {
	case "all":
		rows, err := db.Query("SELECT id FROM teachers")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			rows.Scan(&id)
			teacherIDs = append(teacherIDs, id)
		}
	case "department":
		rows, err := db.Query("SELECT id FROM teachers WHERE department_id=?", req.DepartmentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			rows.Scan(&id)
			teacherIDs = append(teacherIDs, id)
		}
	case "selected":
		teacherIDs = req.TeacherIDs
	}

	// Get project details for email template
	var project Project
	err := db.QueryRow(`
		SELECT id, code, name, email_subject_template, email_body_template, excel_template_filename
		FROM projects WHERE id=?
	`, projectID).Scan(&project.ID, &project.Code, &project.Name, &project.EmailSubjectTemplate,
		&project.EmailBodyTemplate, &project.ExcelTemplateFilename)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project details"})
		return
	}

	// Prepare attachment path
	var attachmentPath string
	if project.ExcelTemplateFilename != "" {
		attachmentPath = filepath.Join("./uploads/templates", project.ExcelTemplateFilename)
	}

	// Insert project members and send emails
	successCount := 0
	for _, tid := range teacherIDs {
		// Get teacher email
		var teacherEmail, teacherName string
		err := db.QueryRow("SELECT name, email FROM teachers WHERE id=?", tid).Scan(&teacherName, &teacherEmail)
		if err != nil {
			log.Printf("Failed to get teacher %d: %v", tid, err)
			continue
		}

		// Replace template variables
		subject := project.EmailSubjectTemplate
		body := project.EmailBodyTemplate
		subject = replaceTemplateVars(subject, teacherName, project.Name)
		body = replaceTemplateVars(body, teacherName, project.Name)

		// Send email
		if err := SendEmail(teacherEmail, subject, body, attachmentPath); err != nil {
			log.Printf("Failed to send email to %s: %v", teacherEmail, err)
			continue
		}

		// Insert/update project member record
		db.Exec(
			"INSERT INTO project_members (project_id, teacher_id, sent_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE sent_at=?",
			projectID, tid, time.Now(), time.Now(),
		)
		successCount++
	}

	// Record dispatch
	db.Exec(
		"INSERT INTO dispatches (project_id, target_type, sent_count) VALUES (?, ?, ?)",
		projectID, req.TargetType, successCount,
	)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Emails dispatched", "count": successCount})
}

func getProjectTracking(c *gin.Context) {
	projectID := c.Param("id")

	query := `
		SELECT 
			t.id, t.name, COALESCE(d.name, '未指定'), 
			pm.current_status, pm.last_reply_at
		FROM project_members pm
		JOIN teachers t ON pm.teacher_id = t.id
		LEFT JOIN departments d ON t.department_id = d.id
		WHERE pm.project_id = ?
	`

	rows, err := db.Query(query, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var records []TrackingRecord
	totalSent := 0
	repliedCount := 0

	for rows.Next() {
		var r TrackingRecord
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

func remindTeachers(c *gin.Context) {
	projectID := c.Param("id")

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

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get project details for reminder template
	var project Project
	err = db.QueryRow(`
		SELECT id, code, name, email_subject_template, email_body_template
		FROM projects WHERE id=?
	`, projectID).Scan(&project.ID, &project.Code, &project.Name,
		&project.EmailSubjectTemplate, &project.EmailBodyTemplate)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project details"})
		return
	}

	successCount := 0
	for rows.Next() {
		var id int
		var name, email string
		rows.Scan(&id, &name, &email)

		// Prepare reminder email content
		subject := "催促提醒: " + project.EmailSubjectTemplate
		body := fmt.Sprintf("尊敬的%s老师：\n\n这是一封催促提醒邮件。\n\n%s\n\n请尽快完成并回复，谢谢！\n\n原邮件内容：\n%s",
			name, project.Name, project.EmailBodyTemplate)

		// Replace template variables
		subject = replaceTemplateVars(subject, name, project.Name)
		body = replaceTemplateVars(body, name, project.Name)

		// Send reminder email
		if err := SendEmail(email, subject, body, ""); err != nil {
			log.Printf("Failed to send reminder to %s (%s): %v", name, email, err)
			continue
		}

		log.Printf("Reminder sent to %s (%s)", name, email)
		successCount++
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Reminders sent", "count": successCount})
}

func fetchProjectEmails(c *gin.Context) {
	projectID := c.Param("id")

	// Convert projectID to int
	var pid int
	if _, err := fmt.Sscanf(projectID, "%d", &pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Process emails for this project
	if err := ProcessEmailsForProject(pid, db); err != nil {
		log.Printf("Failed to process emails for project %d: %v", pid, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"error":   "Failed to fetch emails",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Emails fetched and processed successfully",
	})
}

func aggregateData(c *gin.Context) {
	projectID := c.Param("id")

	// TODO: Implement Excel aggregation logic
	// 1. Get all attachments for this project
	// 2. Parse Excel files
	// 3. Merge data into excel_a_rows or excel_b_rows
	// 4. Generate aggregated Excel file

	log.Printf("Aggregating data for project %s", projectID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Aggregation started"})
}

func downloadAggregated(c *gin.Context) {
	projectID := c.Param("id")

	// TODO: Return aggregated Excel file
	filePath := fmt.Sprintf("./uploads/aggregated/project_%s.xlsx", projectID)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aggregated file not found"})
		return
	}

	c.FileAttachment(filePath, fmt.Sprintf("project_%s_aggregated.xlsx", projectID))
}
