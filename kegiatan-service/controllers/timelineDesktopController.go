package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arkaramadhan/its-vo/common/initializers"
	helper "github.com/arkaramadhan/its-vo/common/utils"
	"github.com/arkaramadhan/its-vo/kegiatan-service/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// GetEventsTimeline retrieves all timeline events
func GetEventsDesktop(c *gin.Context) {
	var events []models.TimelineDesktop
	if err := initializers.DB.Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// CreateEventTimeline creates a new timeline event
func CreateEventDesktop(c *gin.Context) {
	var event models.TimelineDesktop
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Parsing waktu untuk notifikasi
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error loading location"})
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

	// Panggil fungsi SetNotification
	helper.SetNotification(event.Title, startTime, "TimelineDesktop")
	if err := initializers.DB.Create(&event).Error; err != nil {
		log.Printf("Error creating event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

// DeleteEventTimeline deletes a timeline event by ID
func DeleteEventDesktop(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID harus disertakan"})
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID tidak valid"})
		return
	}

	if err := initializers.DB.Where("id = ?", uint(id)).Delete(&models.TimelineDesktop{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func ExportTimelineDesktopHandler(c *gin.Context) {
	var f *excelize.File
	ExportTimelineDesktopToExcel(c, f, "TIMELINE DESKTOP", true)
}

func ExportTimelineDesktopToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	var events_timeline []models.TimelineDesktop
	if err := initializers.DB.Table("kegiatan.timeline_desktops").Find(&events_timeline).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return err
	}

	var excelEvents []helper.ExcelEvent
	for _, event := range events_timeline {
		excelEvents = append(excelEvents, event) // Pastikan `event` adalah tipe yang mengimplementasikan `ExcelEvent`
	}

	config := helper.CalenderConfig{
		SheetName:   "TIMELINE DESKTOP",
		FileName:    "its_report_timelineDesktop.xlsx",
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
