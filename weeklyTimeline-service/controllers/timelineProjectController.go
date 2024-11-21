package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arkaramadhan/its-vo/common/initializers"
	helper "github.com/arkaramadhan/its-vo/common/utils"
	"github.com/arkaramadhan/its-vo/weeklyTimeline-service/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// GetEventsTimeline retrieves all timeline events
func GetEventsProject(c *gin.Context) {
	var events []models.TimelineProject
	helper.FetchAllRecords(initializers.DB, c, &events, "weekly_timeline.timeline_projects", "Gagal mengambil data timeline project")
}

// CreateEventTimeline creates a new timeline event
func CreateEventProject(c *gin.Context) {
	var event models.TimelineProject
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

	// Ubah format parsing sesuai dengan format yang diterima
	startTime, err := time.ParseInLocation("2006-01-02", event.Start, loc) // Ubah format di sini
	if err != nil {
		log.Printf("Error parsing start time: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error parsing start time"})
		return
	}

	// Panggil fungsi SetNotification
	helper.SetNotification(event.Title, startTime, "TimelineProject")

	if err := initializers.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

// DeleteEventTimeline deletes a timeline event by ID
func DeleteEventProject(c *gin.Context) {
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

	// Pastikan ID dikonversi ke tipe data yang sesuai
	if err := initializers.DB.Where("id = ?", uint(id)).Delete(&models.TimelineProject{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Resources

// GetResources retrieves all resources
func GetResourcesProject(c *gin.Context) {
	var resources []models.ResourceProject
	if err := initializers.DB.Find(&resources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Mengembalikan resources sebagai array
	c.JSON(http.StatusOK, resources)
}

// CreateResource creates a new resource
func CreateResourceProject(c *gin.Context) {
	var resource models.ResourceProject
	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := initializers.DB.Create(&resource).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resource)
}

// DeleteResource deletes a resource by ID
func DeleteResourceProject(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" || idParam == "undefined" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID harus disertakan dan valid"})
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID tidak valid"})
		return
	}

	log.Printf("Attempting to delete ResourceProject with ID: %d", id)

	if err := initializers.DB.Where("id = ?", uint(id)).Delete(&models.ResourceProject{}).Error; err != nil {
		log.Printf("Error deleting ResourceProject with ID: %d, error: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	log.Printf("Successfully deleted ResourceProject with ID: %d", id)
	c.Status(http.StatusNoContent)
}

func ExportTimelineProjectHandler(c *gin.Context) {
	var f *excelize.File

	ExportTimelineProjectToExcel(c, f, "TIMELINE PROJECT", true)
}

func ExportTimelineProjectToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	var events_timeline []models.TimelineProject
	if err := initializers.DB.Table("weekly_timeline.timeline_projects").Find(&events_timeline).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return err
	}

	var resources []models.ResourceProject
	if err := initializers.DB.Table("weekly_timeline.resource_projects").Find(&resources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
		SheetName:   "TIMELINE PROJECT",
		FileName:    "its_report_timelineProject.xlsx",
		Events:      excelEvents,
		UseResource: true,
		ResourceMap: resourceMap,
		RowOffset:   0,
		ColOffset:   0,
	}

	if f != nil {
		return helper.ExportCalenderToSheet(f, config)
	} else {
		return helper.ExportCalenderToExcel(c, config)
	}
}