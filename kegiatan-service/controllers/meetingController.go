package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/arkaramadhan/its-vo/common/initializers"
	helper "github.com/arkaramadhan/its-vo/common/utils"
	"github.com/arkaramadhan/its-vo/kegiatan-service/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type MeetingRequest struct {
	ID               uint    `gorm:"primaryKey"`
	Task             *string `json:"task"`
	TindakLanjut     *string `json:"tindak_lanjut"`
	Status           *string `json:"status"`
	UpdatePengerjaan *string `json:"update_pengerjaan"`
	Pic              *string `json:"pic"`
	TanggalTarget    *string `json:"tanggal_target"`
	TanggalActual    *string `json:"tanggal_actual"`
	CreateBy         string  `json:"create_by"`
}

func UploadHandlerMeeting(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/meeting")
}

func GetFilesByIDMeeting(c *gin.Context) {
	helper.GetFilesByID(c)
}

func DeleteFileHandlerMeeting(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/meeting")
}

func DownloadFileHandlerMeeting(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/meeting")
}

func MeetingIndex(c *gin.Context) {
	var meetings []models.Meeting
	helper.FetchAllRecords(initializers.DB, c, &meetings, "kegiatan.meetings", "Gagal mengambil data meeting")
}

func MeetingCreate(c *gin.Context) {

	var requestBody MeetingRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.Status(400)
		c.Error(err) // log the error
		return
	}

	// Add some logging to see what's being received
	log.Println("Received request body:", requestBody)

	// Parse the date string
	tanggal_target, err := time.Parse("2006-01-02", *requestBody.TanggalTarget)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		c.Status(400)
		c.JSON(400, gin.H{"message": "Invalid format tanggal target: " + err.Error()})
		return
	}

	tanggal_actual, err := time.Parse("2006-01-02", *requestBody.TanggalActual)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		c.Status(400)
		c.JSON(400, gin.H{"message": "Invalid format tanggal actual: " + err.Error()})
		return
	}

	requestBody.CreateBy = c.MustGet("username").(string)

	meeting := models.Meeting{
		Task:             requestBody.Task,
		TindakLanjut:     requestBody.TindakLanjut,
		Status:           requestBody.Status,
		UpdatePengerjaan: requestBody.UpdatePengerjaan,
		Pic:              requestBody.Pic,
		TanggalTarget:    &tanggal_target,
		TanggalActual:    &tanggal_actual,
		CreateBy:         requestBody.CreateBy,
	}

	result := initializers.DB.Create(&meeting)

	if result.Error != nil {
		c.Status(400)
		c.JSON(400, gin.H{"error": "Failed to create Meeting: " + result.Error.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "Meeting berhasil dibuat",
	})

}

func MeetingShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var bc models.Meeting
	helper.ShowRecord(c, initializers.DB, id, &bc, "meeting berhasil dilihat", "kegiatan.meetings")
}

func MeetingUpdate(c *gin.Context) {
	var requestBody MeetingRequest
	if err := c.BindJSON(&requestBody); err != nil {
		log.Printf("Error binding JSON: %v", err)
        helper.RespondError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
        return
    }

    var tanggalTarget, tanggalActual *time.Time
    if requestBody.TanggalActual != nil {
        dateFormats := []string{"2006-01-02", "2006-01-02T15:04:05Z07:00", "January 2, 2006", "Jan 2, 2006", "02/01/2006"}
        parsedTanggal, err := helper.ParseFlexibleDate(*requestBody.TanggalActual, dateFormats)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid format tanggal: " + err.Error()})
            return
        }
        tanggalActual = parsedTanggal
    }
    if requestBody.TanggalTarget != nil {
        dateFormats := []string{"2006-01-02", "2006-01-02T15:04:05Z07:00", "January 2, 2006", "Jan 2, 2006", "02/01/2006"}
        parsedTanggal, err := helper.ParseFlexibleDate(*requestBody.TanggalTarget, dateFormats)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid format tanggal: " + err.Error()})
            return
        }
        tanggalTarget = parsedTanggal
    }

    // Assuming you are updating a Memo record
	id := c.Param("id") // or however you get the ID
	var mt models.Meeting
	if err := initializers.DB.First(&mt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "meeting tidak ditemukan"})
        return
    }

    // Update the memo with new data
	if requestBody.Task != nil {
		mt.Task = requestBody.Task
	}
	if requestBody.TindakLanjut != nil {
		mt.TindakLanjut = requestBody.TindakLanjut
	}
	if requestBody.Status != nil {
		mt.Status = requestBody.Status
	}
	if requestBody.UpdatePengerjaan != nil {
		mt.UpdatePengerjaan = requestBody.UpdatePengerjaan
	}
	if requestBody.Pic != nil {
		mt.Pic = requestBody.Pic
	}
	if requestBody.TanggalTarget != nil {
		mt.TanggalTarget = tanggalTarget
	}
	if requestBody.TanggalActual != nil {
		mt.TanggalActual = tanggalActual
	}
	if err := initializers.DB.Save(&mt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update meeting: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "meeting berhasil diupdate"})
}

func MeetingDelete(c *gin.Context) {
	var meeting models.Meeting
	helper.DeleteRecordByID(c, initializers.DB, "kegiatan.meetings", &meeting, "meeting")
}

func CreateExcelMeeting(c *gin.Context) {
	dir := "C:\\excel"
	baseFileName := "its_report_meeting"
	filePath := filepath.Join(dir, baseFileName+".xlsx")

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, append "_new" to the file name
		baseFileName += "_new"
	}

	fileName := baseFileName + ".xlsx"

	// File does not exist, create a new file
	f := excelize.NewFile()

	// Define sheet names
	sheetName := "MEETING"

	// Create sheets and set headers for "MEETING" only
	if sheetName == "MEETING" {
		f.NewSheet(sheetName)
		f.SetCellValue(sheetName, "A1", "TASK")
		f.SetCellValue(sheetName, "B1", "TINDAK LANJUT")
		f.SetCellValue(sheetName, "C1", "STATUS")
		f.SetCellValue(sheetName, "D1", "UPDATE PENGERJAAN")
		f.SetCellValue(sheetName, "E1", "PIC")
		f.SetCellValue(sheetName, "F1", "TANGGAL TARGET")
		f.SetCellValue(sheetName, "G1", "TANGGAL ACTUAL")

	}

	f.SetColWidth(sheetName, "A", "A", 25)
	f.SetColWidth(sheetName, "B", "B", 40)
	f.SetColWidth(sheetName, "C", "C", 17)
	f.SetColWidth(sheetName, "D", "D", 27)
	f.SetColWidth(sheetName, "E", "E", 25)
	f.SetColWidth(sheetName, "F", "F", 20)
	f.SetColWidth(sheetName, "G", "G", 20)
	f.SetRowHeight(sheetName, 1, 35)

	FillColor, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"eba55b"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	f.SetCellStyle(sheetName, "A1", "G1", FillColor)

	wrapstyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,
			Vertical: "center",
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	// Fetch initial data from the database
	var meetings []models.Meeting
	initializers.DB.Find(&meetings)

	f.SetCellStyle(sheetName, "A2", fmt.Sprintf("G%d", len(meetings)+1), wrapstyle)

	// Write initial data to the "MEETING" sheet
	for i, meeting := range meetings {
		tanggalTargetString := meeting.TanggalTarget.Format("2006-01-02")
		tanggalActualString := meeting.TanggalActual.Format("2006-01-02")
		rowNum := i + 2 // Start from the second row (first row is header)

		// Check for nil pointers and use the actual values
		task := ""
		if meeting.Task != nil {
			task = *meeting.Task
		}
		tindakLanjut := ""
		if meeting.TindakLanjut != nil {
			tindakLanjut = *meeting.TindakLanjut
		}
		status := ""
		if meeting.Status != nil {
			status = *meeting.Status
		}
		updatePengerjaan := ""
		if meeting.UpdatePengerjaan != nil {
			updatePengerjaan = *meeting.UpdatePengerjaan
		}
		pic := ""
		if meeting.Pic != nil {
			pic = *meeting.Pic
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), task)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), tindakLanjut)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), status) // Set status value

		// Apply styles based on status
		var styleID int
		switch status {
		case "Done":
			styleID, err = f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Color: "000000",
					Bold:  true,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#5cb85c"},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
			})
		case "On Progress":
			styleID, err = f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Color: "000000",
					Bold:  true,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#f0ad4e"},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
			})
		case "Cancel":
			styleID, err = f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Color: "000000",
					Bold:  true,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#d9534f"},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
			})
		default:
			styleID, err = f.NewStyle(&excelize.Style{
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
			})
		}
		if err != nil {
			fmt.Println(err)
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", rowNum), fmt.Sprintf("C%d", rowNum), styleID)

		// Apply border style to other cells
		borderStyle, err := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
			Alignment: &excelize.Alignment{
				WrapText: true,
			},
		})
		if err != nil {
			fmt.Println(err)
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("A%d", rowNum), borderStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), borderStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", rowNum), fmt.Sprintf("D%d", rowNum), borderStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", rowNum), fmt.Sprintf("E%d", rowNum), borderStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", rowNum), fmt.Sprintf("F%d", rowNum), borderStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), borderStyle)

		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), updatePengerjaan)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), pic)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), tanggalTargetString)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), tanggalActualString)

		// Calculate row height based on content length
		maxContentLength := max(len(task), len(tindakLanjut), len(status), len(updatePengerjaan), len(pic))
		rowHeight := calculateRowHeight(maxContentLength)
		f.SetRowHeight(sheetName, rowNum, rowHeight)
	}

	// Delete the default "Sheet1" sheet
	if err := f.DeleteSheet("Sheet1"); err != nil {
		panic(err) // Handle error jika bukan error "sheet tidak ditemukan"
	}

	// Save the newly created file
	buf, err := f.WriteToBuffer()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error saving file: %v", err)
		return
	}

	// Serve the file to the client
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Writer.Write(buf.Bytes())
}

// Helper function to calculate row height based on content length
func calculateRowHeight(contentLength int) float64 {
	// Define a base height and a multiplier for content length
	baseHeight := 15.0
	multiplier := 0.5
	return baseHeight + (float64(contentLength) * multiplier)
}

// Helper function to find the maximum length among multiple strings
func max(lengths ...int) int {
	maxLength := 0
	for _, length := range lengths {
		if length > maxLength {
			maxLength = length
		}
	}
	return maxLength
}

func ImportExcelMeeting(c *gin.Context) {
	// Mengambil file dari form upload
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error retrieving the file: %v", err)
		return
	}
	defer file.Close()

	// Simpan file sementara jika perlu
	tempFile, err := os.CreateTemp("", "*.xlsx")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating temporary file: %v", err)
		return
	}
	defer os.Remove(tempFile.Name()) // Hapus file sementara setelah selesai

	// Salin file dari request ke file sementara
	if _, err := file.Seek(0, 0); err != nil {
		c.String(http.StatusInternalServerError, "Error seeking file: %v", err)
		return
	}
	if _, err := io.Copy(tempFile, file); err != nil {
		c.String(http.StatusInternalServerError, "Error copying file: %v", err)
		return
	}

	// Buka file Excel dari file sementara
	tempFile.Seek(0, 0) // Reset pointer ke awal file
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file: %v", err)
		return
	}
	defer f.Close()

	// Pilih sheet
	sheetName := "MEETING"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting rows: %v", err)
		return
	}

	// Loop melalui baris dan simpan ke database
	for i, row := range rows {
		if i == 0 {
			// Lewati header baris jika ada
			continue
		}
		if len(row) < 7 {
			// Pastikan ada cukup kolom
			continue
		}
		task := row[0]
		tindakLanjut := row[1]
		status := row[2]
		updatePengerjaan := row[3]
		pic := row[4]
		tanggalTargetString := row[5]
		tanggalActualString := row[6]

		// Parse tanggal
		tanggalTarget, err := time.Parse("2006-01-02", tanggalTargetString)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid date format in row %d: %v", i+1, err)
			return
		}
		tanggalActual, err := time.Parse("2006-01-02", tanggalActualString)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid date format in row %d: %v", i+1, err)
			return
		}

		meeting := models.Meeting{
			Task:             &task,
			TindakLanjut:     &tindakLanjut,
			Status:           &status,
			UpdatePengerjaan: &updatePengerjaan,
			Pic:              &pic,
			TanggalTarget:    &tanggalTarget,
			TanggalActual:    &tanggalActual,
			CreateBy:         c.MustGet("username").(string),
		}

		// Simpan ke database
		if err := initializers.DB.Create(&meeting).Error; err != nil {
			log.Printf("Error saving record from row %d: %v", i+1, err)
			c.String(http.StatusInternalServerError, "Error saving record from row %d: %v", i+1, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport."})
}
