package controllers

import (
	"log"
	"net/http"
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
	helper.GetFilesByID(c, "/app/UploadedFile/meetingschedule")
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
	f := excelize.NewFile()
	ExportMeetingListToExcel(c, f, "MEETING SCHEDULE", true)
}

func ExportMeetingListToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var meetingList []models.MeetingSchedule
	initializers.DB.Table("weekly_timeline.meeting_schedules").Find(&meetingList)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, mt := range meetingList {
		excelData = append(excelData, &mt)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: sheetName,
		Columns: []helper.ExcelColumn{
			{Header: "HARI", Width: 20},
			{Header: "TANGGAL", Width: 20},
			{Header: "PERIHAL", Width: 20},
			{Header: "WAKTU", Width: 20},
			{Header: "SELESAI", Width: 20},
			{Header: "TEMPAT", Width: 20},
			{Header: "STATUS", Width: 20},
			{Header: "PIC", Width: 20},
		},
		Data:         excelData,
		IsSplitSheet: false,
		GetStatus: func(data interface{}) string {
			if mt, ok := data.(*models.MeetingSchedule); ok && mt.Status != nil {
				return *mt.Status
			}
			return "On Progress" // nilai default
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
				"Reschedule": {
					Font: helper.FontBlack,
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"#0275d8"},
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
		fileName := "its_report_meetingList.xlsx"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		if err := f.Write(c.Writer); err != nil {
			return err
		}
	}

	return nil
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

func ImportExcelMeetingList(c *gin.Context) {
	config := helper.ExcelImportConfig{
		SheetName:   "MEETING SCHEDULE",
		MinColumns:  2,
		HeaderRows:  1,
		LogProgress: true,
		ProcessRow: func(row []string, rowIndex int) error {
			// Ambil data dari kolom
			hari := helper.GetColumn(row, 0)
			tanggalStr := helper.GetColumn(row, 1)
			perihal := helper.GetColumn(row, 2)
			waktu := helper.GetColumn(row, 3)
			selesai := helper.GetColumn(row, 4)
			tempat := helper.GetColumn(row, 5)
			status := helper.GetColumn(row, 6)
			pic := helper.GetColumn(row, 7)

			// Parse tanggal
			tanggal, _ := helper.ParseDateWithFormats(tanggalStr)

			// Buat dan simpan memo
			meetingList := models.MeetingSchedule{
				Hari:     &hari,
				Tanggal:  tanggal,
				Perihal:  &perihal,
				Waktu:    &waktu,
				Selesai:  &selesai,
				Tempat:   &tempat,
				Status:   &status,
				Pic:      &pic,
				CreateBy: c.MustGet("username").(string),
			}

			return initializers.DB.Create(&meetingList).Error
		},
	}

	if err := helper.ImportExcelFile(c, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
