package handlers

import (
	"database/sql"
	"net/http"

	"db_intro_backend/db"
	"db_intro_backend/models"

	"github.com/gin-gonic/gin"
)

func GetTeachers(c *gin.Context) {
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
		rows, err = db.DB.Query(query, departmentFilter)
	} else {
		rows, err = db.DB.Query(query)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var teachers []models.Teacher
	for rows.Next() {
		var t models.Teacher
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

func CreateTeacher(c *gin.Context) {
	var t models.Teacher
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.DB.Exec(
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

func UpdateTeacher(c *gin.Context) {
	id := c.Param("id")
	var t models.Teacher
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec(
		"UPDATE teachers SET name=?, email=?, department_id=?, phone=? WHERE id=?",
		t.Name, t.Email, t.DepartmentID, t.Phone, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "updated"})
}

func DeleteTeacher(c *gin.Context) {
	id := c.Param("id")
	_, err := db.DB.Exec("DELETE FROM teachers WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}
