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
func GetEventsRapat(c *gin.Context) {
	var events []models.JadwalRapat
	if err := initializers.DB.Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func CreateEventRapat(c *gin.Context) {
	var event models.JadwalRapat
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	helper.SetNotification(event.Title, startTime, "JadwalRapat") // Panggil fungsi SetNotification

	if err := initializers.DB.Create(&event).Error; err != nil {
		log.Printf("Error creating event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func DeleteEventRapat(c *gin.Context) {
	id := c.Param("id") // Menggunakan c.Param jika UUID dikirim sebagai bagian dari URL
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID harus disertakan"})
		return
	}
	if err := initializers.DB.Where("id = ?", id).Delete(&models.JadwalRapat{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func ExportJadwalRapatHandler(c *gin.Context) {
	var f *excelize.File
	ExportJadwalRapatToExcel(c, f, "JADWAL RAPAT", true)
}

func ExportJadwalRapatToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	var events []models.JadwalRapat
	if err := initializers.DB.Table("kegiatan.jadwal_rapats").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	var excelEvents []helper.ExcelEvent
	for _, event := range events {
		excelEvents = append(excelEvents, event) // Pastikan `event` adalah tipe yang mengimplementasikan `ExcelEvent`
	}

	config := helper.CalenderConfig{
		SheetName:   "JADWAL RAPAT",
		FileName:    "jadwal_rapat.xlsx",
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
