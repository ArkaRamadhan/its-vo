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

// Create a new event
func GetEventsCuti(c *gin.Context) {
	var events []models.JadwalCuti
	if err := initializers.DB.Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// Example of using generated UUID
func CreateEventCuti(c *gin.Context) {
	var event models.JadwalCuti
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Set notification menggunakan fungsi dari notificationController
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Error loading location: %v", err)
		return
	}

	var startTime time.Time
	if event.AllDay {
		// Jika AllDay = true, set waktu ke tengah malam
		startTime, err = time.ParseInLocation("2006-01-02T15:04:05", event.Start+"T00:00:00", loc)
	} else {
		// Jika tidak, parse dengan format RFC3339
		startTime, err = time.ParseInLocation(time.RFC3339, event.Start, loc)
	}

	if err != nil {
		log.Printf("Error parsing start time: %v", err)
		return
	}

	helper.SetNotification(event.Title, startTime, "JadwalCuti") // Panggil fungsi SetNotification
	if err := initializers.DB.Create(&event).Error; err != nil {
		log.Printf("Error creating event: %v", err) // Add this line
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func DeleteEventCuti(c *gin.Context) {
	id := c.Param("id") // Menggunakan c.Param jika UUID dikirim sebagai bagian dari URL
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID harus disertakan"})
		return
	}
	if err := initializers.DB.Where("id = ?", id).Delete(&models.JadwalCuti{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func ExportJadwalCutiHandler(c *gin.Context) {
	var f *excelize.File
	ExportJadwalCutiToExcel(c, f, "JADWAL CUTI", true)
}

func ExportJadwalCutiToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	var events []models.JadwalCuti
	if err := initializers.DB.Table("kegiatan.jadwal_cutis").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return err
	}

	var excelEvents []helper.ExcelEvent
	for _, event := range events {
		excelEvents = append(excelEvents, event) // Pastikan `event` adalah tipe yang mengimplementasikan `ExcelEvent`
	}

	config := helper.CalenderConfig{
		SheetName:   "JADWAL CUTI",
		FileName:    "jadwal_cuti.xlsx",
		Events:      excelEvents,
		UseResource: false,
		RowOffset:   0,
		ColOffset:   0,
	}

	if f != nil {
		return helper.ExportCalenderToSheet(f, config)
	} else {
		err := helper.ExportCalenderToExcel(c, config)
		if err != nil {
			c.String(http.StatusInternalServerError, "Gagal mengekspor data ke Excel: "+err.Error())
			return err
		}
	}
	return nil
}
