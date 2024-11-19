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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// CreateEventTimeline creates a new timeline event
func CreateEventDesktop(c *gin.Context) {
	var event models.TimelineDesktop
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parsing waktu untuk notifikasi
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error loading location"})
		return
	}

	// Ubah format parsing sesuai dengan format yang diterima
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", event.Start, loc) // Ubah format di sini
	if err != nil {
		log.Printf("Error parsing start time: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error parsing start time"})
		return
	}

	// Panggil fungsi SetNotification
	helper.SetNotification(event.Title, startTime, "TimelineWallpaperDesktop")

	if err := initializers.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

// Resources

// GetResources retrieves all resources
func GetResourcesDesktop(c *gin.Context) {
	var resources []models.ResourceDesktop
	if err := initializers.DB.Find(&resources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resources)
}

// CreateResource creates a new resource
func CreateResourceDesktop(c *gin.Context) {
	var resource models.ResourceDesktop
	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := initializers.DB.Create(&resource).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resource)
}

// DeleteResource deletes a resource by ID
func DeleteResourceDesktop(c *gin.Context) {
	id := c.Param("id")
	if err := initializers.DB.Where("id = ?", id).Delete(&models.ResourceDesktop{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	var resources []models.ResourceDesktop
	if err := initializers.DB.Table("kegiatan.resource_desktops").Find(&resources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	resourceMap := make(map[uint]string)
	for _, resource := range resources {
		resourceMap[resource.ID] = resource.Name
	}

	var excelEvents []helper.ExcelEvent
	for _, event := range events_timeline {
		excelEvents = append(excelEvents, event) // Pastikan `event` adalah tipe yang mengimplementasikan `ExcelEvent`
	}

	config := helper.CalenderConfig{
		SheetName:   "TIMELINE DESKTOP",
		FileName:    "its_report_timelineDesktop.xlsx",
		Events:      excelEvents,
		UseResource: true,
		ResourceMap: resourceMap,
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
