package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"db_intro_backend/db"
	"db_intro_backend/models"

	"github.com/xuri/excelize/v2"
)

var (
	ErrNoExcelAttachments = errors.New("no Excel attachments found for this project")
)

type ExcelService struct{}

func NewExcelService() *ExcelService {
	return &ExcelService{}
}

func (s *ExcelService) AggregateProjectExcel(projectID int) (string, int, int, error) {
	attachments, err := s.fetchProjectExcelAttachments(projectID)
	if err != nil {
		return "", 0, 0, err
	}
	if len(attachments) == 0 {
		return "", 0, 0, ErrNoExcelAttachments
	}

	if err := os.MkdirAll("./uploads/aggregated", 0755); err != nil {
		return "", 0, 0, fmt.Errorf("failed to prepare aggregated directory: %w", err)
	}

	outputPath := s.AggregatedFilePath(strconv.Itoa(projectID))
	book := excelize.NewFile()
	defer book.Close()

	sheetName := "Aggregated"
	book.SetSheetName(book.GetSheetName(0), sheetName)
	metaHeaders := []string{"Teacher", "Email", "Source File"}
	headerRow := append([]string{}, metaHeaders...)
	headerWritten := false
	dataColumns := 0
	rowCursor := 2
	processedAttachments := 0
	appendedRows := 0

	for _, att := range attachments {
		if !s.isExcelFile(att.OriginalName) && !s.isExcelFile(att.StoredPath) {
			continue
		}

		if _, err := os.Stat(att.StoredPath); err != nil {
			log.Printf("Attachment missing (%s): %v", att.StoredPath, err)
			continue
		}

		file, err := excelize.OpenFile(att.StoredPath)
		if err != nil {
			log.Printf("Failed to open attachment %s: %v", att.StoredPath, err)
			continue
		}

		sheets := file.GetSheetList()
		if len(sheets) == 0 {
			file.Close()
			continue
		}

		rows, err := file.GetRows(sheets[0])
		file.Close()
		if err != nil {
			log.Printf("Failed to read rows from %s: %v", att.StoredPath, err)
			continue
		}

		headerIdx, headerCells := s.firstNonEmptyRow(rows)
		if headerIdx == -1 {
			continue
		}

		if !headerWritten {
			headerRow = append(headerRow[:0], metaHeaders...)
			headerRow = append(headerRow, headerCells...)
			dataColumns = len(headerCells)
			headCopy := make([]string, len(headerRow))
			copy(headCopy, headerRow)
			book.SetSheetRow(sheetName, "A1", &headCopy)
			headerWritten = true
		} else if len(headerCells) > dataColumns {
			headerRow = append(headerRow, headerCells[dataColumns:]...)
			rowCopy := make([]string, len(headerRow))
			copy(rowCopy, headerRow)
			book.SetSheetRow(sheetName, "A1", &rowCopy)
			dataColumns = len(headerCells)
		}

		dataRows := rows[headerIdx+1:]
		if len(dataRows) == 0 {
			processedAttachments++
			continue
		}

		teacherName := att.TeacherName
		if teacherName == "" {
			teacherName = "未匹配教师"
		}

		teacherEmail := att.TeacherEmail

		for _, dataRow := range dataRows {
			if s.rowIsEmpty(dataRow) {
				continue
			}

			if len(dataRow) > dataColumns {
				extra := len(dataRow) - dataColumns
				for i := 0; i < extra; i++ {
					headerRow = append(headerRow, fmt.Sprintf("ExtraCol_%d", dataColumns+i+1))
				}
				dataColumns = len(dataRow)
				rowCopy := make([]string, len(headerRow))
				copy(rowCopy, headerRow)
				book.SetSheetRow(sheetName, "A1", &rowCopy)
			}

			if len(dataRow) < dataColumns {
				padding := make([]string, dataColumns-len(dataRow))
				dataRow = append(dataRow, padding...)
			}

			metaValues := []string{teacherName, teacherEmail, att.OriginalName}
			rowValues := append(metaValues, dataRow...)
			rowCopy := make([]string, len(rowValues))
			copy(rowCopy, rowValues)
			book.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowCursor), &rowCopy)
			rowCursor++
			appendedRows++
		}

		processedAttachments++
	}

	if !headerWritten {
		return "", 0, 0, ErrNoExcelAttachments
	}

	if err := book.SaveAs(outputPath); err != nil {
		return "", 0, 0, fmt.Errorf("failed to save aggregated workbook: %w", err)
	}

	return outputPath, processedAttachments, appendedRows, nil
}

func (s *ExcelService) fetchProjectExcelAttachments(projectID int) ([]models.AttachmentMeta, error) {
	rows, err := db.DB.Query(`
		SELECT a.stored_path, a.original_filename, COALESCE(t.name, ''), COALESCE(t.email, '')
		FROM attachments a
		LEFT JOIN teachers t ON a.teacher_id = t.id
		WHERE a.project_id = ?
		ORDER BY a.created_at ASC, a.id ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []models.AttachmentMeta
	for rows.Next() {
		var att models.AttachmentMeta
		if err := rows.Scan(&att.StoredPath, &att.OriginalName, &att.TeacherName, &att.TeacherEmail); err != nil {
			continue
		}
		attachments = append(attachments, att)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return attachments, nil
}

func (s *ExcelService) firstNonEmptyRow(rows [][]string) (int, []string) {
	for idx, row := range rows {
		if !s.rowIsEmpty(row) {
			return idx, row
		}
	}
	return -1, nil
}

func (s *ExcelService) rowIsEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func (s *ExcelService) isExcelFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".xlsx", ".xls", ".xlsm", ".xlsb":
		return true
	default:
		return false
	}
}

func (s *ExcelService) AggregatedFilePath(projectID string) string {
	return fmt.Sprintf("./uploads/aggregated/project_%s.xlsx", projectID)
}
