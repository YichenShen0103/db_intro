package handlers

import (
	"net/http"

	"db_intro_backend/db"
	"db_intro_backend/models"

	"github.com/gin-gonic/gin"
)

func GetDepartments(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, name, code, created_at FROM departments")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var d models.Department
		if err := rows.Scan(&d.ID, &d.Name, &d.Code, &d.CreatedAt); err != nil {
			continue
		}
		departments = append(departments, d)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": departments})
}
