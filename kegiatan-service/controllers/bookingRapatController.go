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
func GetEventsBookingRapat(c *gin.Context) {
	var events []models.BookingRapat
	// Tambahkan filter untuk tidak menampilkan event dengan status "pending"
	if err := initializers.DB.Where("status != ?", "pending").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// Example of using generated UUID
func CreateEventBookingRapat(c *gin.Context) {
	var event models.BookingRapat
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simpan event ke database terlebih dahulu
	if err := initializers.DB.Create(&event).Error; err != nil {
		log.Printf("Error creating event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log untuk memeriksa data yang diterima
	log.Printf("Event Start: %s, Event End: %s", event.Start, event.End)

	// Set notification menggunakan fungsi dari notificationController
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Error loading location: %v", err)
		return
	}

	var startTime time.Time
	if event.AllDay {
		startTime, err = time.ParseInLocation("2006-01-02T15:04:05", event.Start+"T00:00:00", loc)
	} else {
		startTime, err = time.ParseInLocation(time.RFC3339, event.Start, loc)
	}

	if err != nil {
		log.Printf("Error parsing start time: %v", err)
		return
	}

	// Panggil fungsi SetNotification setelah event berhasil disimpan
	helper.SetNotification(event.Title, startTime, "BookingRapat")

	// Cek bentrok, kecualikan event yang sedang dibuat
	var conflictingEvents []models.BookingRapat
	if err := initializers.DB.Where("id != ? AND start < ? AND \"end\" > ?", event.ID, event.End, event.Start).Find(&conflictingEvents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log untuk memeriksa hasil query
	log.Printf("Jumlah jadwal bentrok: %d", len(conflictingEvents))
	for _, conflict := range conflictingEvents {
		log.Printf("Bentrok dengan %s: Start: %s, End: %s", conflict.Title, conflict.Start, conflict.End)
	}

	// Atur status berdasarkan bentrok
	if len(conflictingEvents) > 0 {
		event.Status = "pending"
	} else {
		event.Status = "acc"
	}

	// Update status event di database
	if err := initializers.DB.Save(&event).Error; err != nil {
		log.Printf("Error updating event status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

func DeleteEventBookingRapat(c *gin.Context) {
	id := c.Param("id") // Menggunakan c.Param jika UUID dikirim sebagai bagian dari URL
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID harus disertakan"})
		return
	}
	if err := initializers.DB.Where("id = ?", id).Delete(&models.BookingRapat{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func ExportBookingRapatHandler(c *gin.Context) {
	var f *excelize.File
	ExportBookingRapatToExcel(c, f, "BOOKING RAPAT", true)
}

func ExportBookingRapatToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	var events []models.BookingRapat
	if err := initializers.DB.Where("status = ?", "acc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	var excelEvents []helper.ExcelEvent
	for _, event := range events {
		excelEvents = append(excelEvents, event) // Pastikan `event` adalah tipe yang mengimplementasikan `ExcelEvent`
	}

	config := helper.CalenderConfig{
		SheetName:   "BOOKING RAPAT",
		FileName:    "bookingRapat.xlsx",
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
