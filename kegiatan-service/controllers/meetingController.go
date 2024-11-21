package controllers

import (
	"log"
	"net/http"
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
	helper.GetFilesByID(c, "/app/UploadedFile/meeting")
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

func ExportMeetingHandler(c *gin.Context) {
	f := excelize.NewFile()

	ExportMeetingToExcel(c, f, "MEETING", true)
}

func ExportMeetingToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var meeting []models.Meeting
	initializers.DB.Table("kegiatan.meetings").Find(&meeting)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, mt := range meeting {
		excelData = append(excelData, &mt)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "MEETING",
		Columns: []helper.ExcelColumn{
			{Header: "TASK", Width: 24},
			{Header: "TINDAK LANJUT", Width: 40},
			{Header: "STATUS", Width: 27},
			{Header: "UPDATE PENGERJAAN", Width: 27},
			{Header: "PIC", Width: 20},
			{Header: "TANGGAL TARGET", Width: 20},
			{Header: "TANGGAL ACTUAL", Width: 20},
		},
		Data:         excelData,
		IsSplitSheet: false,
		GetStatus: func(data interface{}) string {
			if mt, ok := data.(*models.Meeting); ok && mt.Status != nil {
				return *mt.Status
			}
			return "Pending" // nilai default
		},
		CustomStyles: &helper.CustomStyles{
			StatusStyles: map[string]*excelize.Style{
				"Done": {
					Font: helper.FontBlack,
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"#5cb85c"},
						Pattern: 1,
					},
					Alignment: helper.CenterAlignment,
					Border:    helper.BorderBlack,
				},
				"On Progress": {
					Font: helper.FontBlack,
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"#f0ad4e"},
						Pattern: 1,
					},
					Alignment: helper.CenterAlignment,
					Border:    helper.BorderBlack,
				},
				"Cancel": {
					Font: helper.FontBlack,
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"#d9534f"},
						Pattern: 1,
					},
					Alignment: helper.CenterAlignment,
					Border:    helper.BorderBlack,
				},
			},
			DefaultCellStyle: &excelize.Style{
				Border:    helper.BorderBlack,
				Alignment: helper.WrapAlignment,
			},
		},
	}

	if f != nil {
		helper.ExportToSheet(f, config)
	} else {
		helper.ExportToExcel(config)
	}

	if isStandAlone {
		fileName := "its_report_suratkeluar.xlsx"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		if err := f.Write(c.Writer); err != nil {
			return err
		}
	}

	return nil
}

// func ImportExcelMeeting(c *gin.Context) {
// 	// Mengambil file dari form upload
// 	file, _, err := c.Request.FormFile("file")
// 	if err != nil {
// 		c.String(http.StatusBadRequest, "Error retrieving the file: %v", err)
// 		return
// 	}
// 	defer file.Close()

// 	// Simpan file sementara jika perlu
// 	tempFile, err := os.CreateTemp("", "*.xlsx")
// 	if err != nil {
// 		c.String(http.StatusInternalServerError, "Error creating temporary file: %v", err)
// 		return
// 	}
// 	defer os.Remove(tempFile.Name()) // Hapus file sementara setelah selesai

// 	// Salin file dari request ke file sementara
// 	if _, err := file.Seek(0, 0); err != nil {
// 		c.String(http.StatusInternalServerError, "Error seeking file: %v", err)
// 		return
// 	}
// 	if _, err := io.Copy(tempFile, file); err != nil {
// 		c.String(http.StatusInternalServerError, "Error copying file: %v", err)
// 		return
// 	}

// 	// Buka file Excel dari file sementara
// 	tempFile.Seek(0, 0) // Reset pointer ke awal file
// 	f, err := excelize.OpenFile(tempFile.Name())
// 	if err != nil {
// 		c.String(http.StatusInternalServerError, "Error opening file: %v", err)
// 		return
// 	}
// 	defer f.Close()

// 	// Pilih sheet
// 	sheetName := "MEETING"
// 	rows, err := f.GetRows(sheetName)
// 	if err != nil {
// 		c.String(http.StatusInternalServerError, "Error getting rows: %v", err)
// 		return
// 	}

// 	// Loop melalui baris dan simpan ke database
// 	for i, row := range rows {
// 		if i == 0 {
// 			// Lewati header baris jika ada
// 			continue
// 		}
// 		if len(row) < 7 {
// 			// Pastikan ada cukup kolom
// 			continue
// 		}
// 		task := row[0]
// 		tindakLanjut := row[1]
// 		status := row[2]
// 		updatePengerjaan := row[3]
// 		pic := row[4]
// 		tanggalTargetString := row[5]
// 		tanggalActualString := row[6]

// 		// Parse tanggal
// 		tanggalTarget, err := time.Parse("2006-01-02", tanggalTargetString)
// 		if err != nil {
// 			c.String(http.StatusBadRequest, "Invalid date format in row %d: %v", i+1, err)
// 			return
// 		}
// 		tanggalActual, err := time.Parse("2006-01-02", tanggalActualString)
// 		if err != nil {
// 			c.String(http.StatusBadRequest, "Invalid date format in row %d: %v", i+1, err)
// 			return
// 		}

// 		meeting := models.Meeting{
// 			Task:             &task,
// 			TindakLanjut:     &tindakLanjut,
// 			Status:           &status,
// 			UpdatePengerjaan: &updatePengerjaan,
// 			Pic:              &pic,
// 			TanggalTarget:    &tanggalTarget,
// 			TanggalActual:    &tanggalActual,
// 			CreateBy:         c.MustGet("username").(string),
// 		}

// 		// Simpan ke database
// 		if err := initializers.DB.Create(&meeting).Error; err != nil {
// 			log.Printf("Error saving record from row %d: %v", i+1, err)
// 			c.String(http.StatusInternalServerError, "Error saving record from row %d: %v", i+1, err)
// 			return
// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport."})
// }

func ImportExcelMeeting(c *gin.Context) {
	config := helper.ExcelImportConfig{
		SheetName:   "MEETING",
		MinColumns:  2,
		HeaderRows:  1,
		LogProgress: true,
		ProcessRow: func(row []string, rowIndex int) error {
			// Ambil data dari kolom
			task := helper.GetColumn(row, 0)
			tindakLanjut := helper.GetColumn(row, 1)
			status := helper.GetColumn(row, 2)
			updatePengerjaan := helper.GetColumn(row, 3)
			pic := helper.GetColumn(row, 4)
			tanggalTargetStr := helper.GetColumn(row, 5)
			tanggalActualStr := helper.GetColumn(row, 6)

			// Parse tanggal
			tanggalTarget, _ := helper.ParseDateWithFormats(tanggalTargetStr)
			tanggalActual, _ := helper.ParseDateWithFormats(tanggalActualStr)

			// Buat dan simpan memo
			meeting := models.Meeting{
				Task:             &task,
				TindakLanjut:     &tindakLanjut,
				Status:           &status,
				UpdatePengerjaan: &updatePengerjaan,
				Pic:              &pic,
				TanggalTarget:    tanggalTarget,
				TanggalActual:    tanggalActual,
				CreateBy:         c.MustGet("username").(string),
			}

			return initializers.DB.Create(&meeting).Error
		},
	}

	if err := helper.ImportExcelFile(c, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
