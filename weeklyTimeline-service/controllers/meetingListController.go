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
	"github.com/arkaramadhan/its-vo/weeklyTimeline-service/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type MeetingListRequest struct {
	ID       uint    `gorm:"primaryKey"`
	Hari     *string `json:"hari"`
	Tanggal  *string `json:"tanggal"`
	Perihal  *string `json:"perihal"`
	Waktu    *string `json:"waktu"`
	Selesai  *string `json:"selesai"`
	Tempat   *string `json:"tempat"`
	Pic      *string `json:"pic"`
	Status   *string `json:"status"`
	CreateBy string  `json:"create_by"`
	Info     string  `json:"info"`
	Color    string  `json:"color"`
}

func UploadHandlerMeetingList(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/meetingschedule")
}

func GetFilesByIDMeetingList(c *gin.Context) {
	helper.GetFilesByID(c)
}

func DeleteFileHandlerMeetingList(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/meetingschedule")
}

func DownloadFileHandlerMeetingList(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/meetingschedule")
}

func MeetingListIndex(c *gin.Context) {
	var meetings []models.MeetingSchedule
	helper.FetchAllRecords(initializers.DB, c, &meetings, "weekly_timeline.meeting_schedules", "Gagal mengambil data meeting schedule")
}

func MeetingListCreate(c *gin.Context) {

	var requestBody MeetingListRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.Status(400)
		c.Error(err) // log the error
		return
	}

	// Add some logging to see what's being received
	log.Println("Received request body:", requestBody)

	// Parse the date string
	tanggal, err := time.Parse("2006-01-02", *requestBody.Tanggal)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		c.Status(400)
		c.JSON(400, gin.H{"message": "Invalid format tanggal: " + err.Error()})
		return
	}

	requestBody.CreateBy = c.MustGet("username").(string)

	meetingList := models.MeetingSchedule{
		Hari:     requestBody.Hari,
		Tanggal:  &tanggal,
		Perihal:  requestBody.Perihal,
		Waktu:    requestBody.Waktu,
		Selesai:  requestBody.Selesai,
		Tempat:   requestBody.Tempat,
		Pic:      requestBody.Pic,
		Status:   requestBody.Status,
		CreateBy: requestBody.CreateBy,
		Color:    requestBody.Color,
	}

	result := initializers.DB.Create(&meetingList)

	if result.Error != nil {
		c.Status(400)
		c.JSON(400, gin.H{"message": "gagal membuat Meeting: " + result.Error.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "Meeting berhasil dibuat",
	})

}

func MeetingListShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var bc models.MeetingSchedule
	helper.ShowRecord(c, initializers.DB, id, &bc, "meeting schedule berhasil dilihat", "weekly_timeline.meeting_schedules")
}

func MeetingListUpdate(c *gin.Context) {
	var requestBody MeetingListRequest
	if err := c.BindJSON(&requestBody); err != nil {
		log.Printf("Error binding JSON: %v", err)
        helper.RespondError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
        return
    }

    var tanggal *time.Time
    if requestBody.Tanggal != nil {
        dateFormats := []string{"2006-01-02", "2006-01-02T15:04:05Z07:00", "January 2, 2006", "Jan 2, 2006", "02/01/2006"}
        parsedTanggal, err := helper.ParseFlexibleDate(*requestBody.Tanggal, dateFormats)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid format tanggal: " + err.Error()})
            return
        }
        tanggal = parsedTanggal
    }

    // Assuming you are updating a Memo record
	id := c.Param("id") // or however you get the ID
	var ml models.MeetingSchedule
	if err := initializers.DB.First(&ml, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "meeting list tidak ditemukan"})
        return
    }

    // Update the memo with new data
	if requestBody.Hari != nil {
		ml.Hari = requestBody.Hari
	}
	if tanggal != nil {
		ml.Tanggal = tanggal
	}
	if requestBody.Perihal != nil {
		ml.Perihal = requestBody.Perihal
	}
	if requestBody.Waktu != nil {
		ml.Waktu = requestBody.Waktu
	}
	if requestBody.Selesai != nil {
		ml.Selesai = requestBody.Selesai
	}
	if requestBody.Tempat != nil {
		ml.Tempat = requestBody.Tempat
	}
	if requestBody.Pic != nil {
		ml.Pic = requestBody.Pic
	}
	if requestBody.Status != nil {
		ml.Status = requestBody.Status
	}
	if requestBody.Color != "" {
		ml.Color = requestBody.Color
	}
	if err := initializers.DB.Save(&ml).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update meeting list: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "meeting list berhasil diupdate"})
}

func MeetingListDelete(c *gin.Context) {
	var meetingList models.MeetingSchedule
	helper.DeleteRecordByID(c, initializers.DB, "weekly_timeline.meeting_schedules", &meetingList, "meeting schedule")
}

func CreateExcelMeetingList(c *gin.Context) {
	dir := "C:\\excel"
	baseFileName := "its_report_meetingSchedule"
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
	sheetName := "MEETING SCHEDULE"

	// Create sheets and set headers for "MEETING SCHEDULE" only
	f.NewSheet(sheetName)
	f.SetCellValue(sheetName, "A1", "Hari")
	f.SetCellValue(sheetName, "B1", "Tanggal")
	f.SetCellValue(sheetName, "C1", "Perihal")
	f.SetCellValue(sheetName, "D1", "Waktu")
	f.SetCellValue(sheetName, "E1", "Selesai")
	f.SetCellValue(sheetName, "F1", "Tempat")
	f.SetCellValue(sheetName, "G1", "Status")
	f.SetCellValue(sheetName, "H1", "PIC")

	f.SetColWidth(sheetName, "A", "H", 20)
	f.SetRowHeight(sheetName, 1, 20)

	styleHeader, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#6EB6F8"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return
	}

	f.SetCellStyle(sheetName, "A1", "H1", styleHeader)

	// Fetch initial data from the database
	var meetingList []models.MeetingSchedule
	initializers.DB.Find(&meetingList)

	// Definisikan gaya untuk border
	styleAll, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Error membuat gaya: %v", err)
		return
	}

	// Definisikan gaya untuk status yang berbeda
	styleDone, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "000000", Bold: true},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#5cb85c"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	styleCancel, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "000000", Bold: true},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#d9534f"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	styleReschedule, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "000000", Bold: true},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#0275d8"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	styleOnProgress, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "000000", Bold: true},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#f0ad4e"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Tulis data awal ke lembar "MEETING SCHEDULE"
	meetingListSheetName := "MEETING SCHEDULE"
	for i, meetingList := range meetingList {
		rowNum := i + 2 // Start from the second row (first row is header)

		// Format Hari dan Tanggal
		formattedDate := ""
		if meetingList.Tanggal != nil {
			dayInEnglish := meetingList.Tanggal.Format("Monday") // Dapatkan nama hari dalam bahasa Inggris
			dayInIndonesian := hariIndonesia(dayInEnglish)       // Konversi ke bahasa Indonesia
			formattedDate = dayInIndonesian + ", " + meetingList.Tanggal.Format("2006-01-02")
		}
		f.SetCellValue(meetingListSheetName, fmt.Sprintf("B%d", rowNum), formattedDate)

		// Periksa dan atur nilai Hari
		if meetingList.Hari != nil {
			f.SetCellValue(meetingListSheetName, fmt.Sprintf("A%d", rowNum), *meetingList.Hari)
		} else {
			f.SetCellValue(meetingListSheetName, fmt.Sprintf("A%d", rowNum), "")
		}

		f.SetCellValue(meetingListSheetName, fmt.Sprintf("C%d", rowNum), *meetingList.Perihal)

		// Handle Waktu
		if meetingList.Waktu != nil {
			f.SetCellValue(meetingListSheetName, fmt.Sprintf("D%d", rowNum), *meetingList.Waktu)
		} else {
			f.SetCellValue(meetingListSheetName, fmt.Sprintf("D%d", rowNum), "")
		}

		// Handle Selesai
		if meetingList.Selesai != nil {
			f.SetCellValue(meetingListSheetName, fmt.Sprintf("E%d", rowNum), *meetingList.Selesai)
		} else {
			f.SetCellValue(meetingListSheetName, fmt.Sprintf("E%d", rowNum), "")
		}

		if meetingList.Tempat != nil {
			f.SetCellValue(meetingListSheetName, fmt.Sprintf("F%d", rowNum), *meetingList.Tempat)
		} else {
			f.SetCellValue(meetingListSheetName, fmt.Sprintf("F%d", rowNum), "")
		}

		f.SetCellValue(meetingListSheetName, fmt.Sprintf("G%d", rowNum), *meetingList.Status)
		f.SetCellValue(meetingListSheetName, fmt.Sprintf("H%d", rowNum), *meetingList.Pic)

		// Terapkan gaya border untuk semua sel
		for col := 'A'; col <= 'H'; col++ {
			cellName := fmt.Sprintf("%c%d", col, rowNum)
			f.SetCellStyle(meetingListSheetName, cellName, cellName, styleAll)
		}

		// Terapkan gaya khusus untuk status
		switch *meetingList.Status {
		case "Done":
			f.SetCellStyle(meetingListSheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), styleDone)
		case "Cancel":
			f.SetCellStyle(meetingListSheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), styleCancel)
		case "Reschedule":
			f.SetCellStyle(meetingListSheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), styleReschedule)
		case "On Progress":
			f.SetCellStyle(meetingListSheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), styleOnProgress)
		}
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

func ImportExcelMeetingList(c *gin.Context) {
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
	sheetName := "MEETING SCHEDULE"
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
		if len(row) < 4 {
			// Pastikan ada cukup kolom
			continue
		}
		hari := row[0]
		tanggal := row[1]
		perihal := row[2]
		waktu := row[3]
		selesai := row[4]
		tempat := row[5]
		status := row[6]
		pic := row[7]

		// Parse tanggal
		tanggalString, err := time.Parse("2006-01-02", tanggal)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid date format in row %d: %v", i+1, err)
			return
		}

		meetingList := models.MeetingSchedule{
			Hari:     &hari,
			Tanggal:  &tanggalString,
			Perihal:  &perihal,
			Waktu:    &waktu,
			Selesai:  &selesai,
			Tempat:   &tempat,
			Status:   &status,
			Pic:      &pic,
			CreateBy: c.MustGet("username").(string),
		}

		// Simpan ke database
		if err := initializers.DB.Create(&meetingList).Error; err != nil {
			log.Printf("Error saving record from row %d: %v", i+1, err)
			c.String(http.StatusInternalServerError, "Error saving record from row %d: %v", i+1, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}

func hariIndonesia(day string) string {
	days := map[string]string{
		"Monday":    "Senin",
		"Tuesday":   "Selasa",
		"Wednesday": "Rabu",
		"Thursday":  "Kamis",
		"Friday":    "Jumat",
		"Saturday":  "Sabtu",
		"Sunday":    "Minggu",
	}
	return days[day]
}
