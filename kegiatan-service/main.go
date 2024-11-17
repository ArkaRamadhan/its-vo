package main

import (
	"log"
	"os"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/middleware"
	"github.com/arkaramadhan/its-vo/common/utils"
	"github.com/arkaramadhan/its-vo/kegiatan-service/controllers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

)

func init() {
	initializers.LoadEnvVariables()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set in the environment variables")
	}

	err := initializers.ConnectToDB("kegiatan")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func main() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(middleware.CORS())

	// ********** Middleware ********** //
	r.Use(middleware.TokenAuthMiddleware())
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// ********** Route Timeline Desktop ********** //
	r.GET("/timelineDesktop", controllers.GetEventsDesktop)
	r.POST("/timelineDesktop", controllers.CreateEventDesktop)
	r.DELETE("/timelineDesktop/:id", controllers.DeleteEventDesktop)
	r.GET("/resourceDesktop", controllers.GetResourcesDesktop)
	r.POST("/resourceDesktop", controllers.CreateResourceDesktop)
	r.DELETE("/resourceDesktop/:id", controllers.DeleteResourceDesktop)
	r.GET("/exportTimelineDesktop", controllers.ExportTimelineDesktopToExcel)

	// ********** Route Booking Rapat ********** //
	r.GET("/booking-rapat", controllers.GetEventsBookingRapat)
	r.POST("/booking-rapat", controllers.CreateEventBookingRapat)
	r.DELETE("/booking-rapat/:id", controllers.DeleteEventBookingRapat)
	r.GET("/exportBookingRapat", controllers.ExportBookingRapatToExcel)

	// ********** Route Jadwal Rapat ********** //
	r.GET("/jadwal-rapat", controllers.GetEventsRapat)
	r.POST("/jadwal-rapat", controllers.CreateEventRapat)
	r.DELETE("/jadwal-rapat/:id", controllers.DeleteEventRapat)
	r.GET("/exportRapat", controllers.ExportJadwalRapatToExcel)

	// ********** Route Jadwal Cuti ********** //
	r.GET("/jadwal-cuti", controllers.GetEventsCuti)
	r.POST("/jadwal-cuti", controllers.CreateEventCuti)
	r.DELETE("/jadwal-cuti/:id", controllers.DeleteEventCuti)
	r.GET("/exportCuti", controllers.ExportJadwalCutiToExcel)

	// ********** Route Meeting ********** //
	r.GET("/meetings", controllers.MeetingIndex)
	r.POST("/meetings", controllers.MeetingCreate)
	r.GET("/meetings/:id", controllers.MeetingShow)
	r.PUT("/meetings/:id", controllers.MeetingUpdate)
	r.DELETE("/meetings/:id", controllers.MeetingDelete)
	r.GET("/exportMeeting", controllers.CreateExcelMeeting)
	r.POST("/uploadMeeting", controllers.ImportExcelMeeting)

	r.POST("/uploadFileMeeting", controllers.UploadHandlerMeeting)
	r.GET("/downloadMeeting/:id/:filename", controllers.DownloadFileHandlerMeeting)
	r.DELETE("/deleteMeeting/:id/:filename", controllers.DeleteFileHandlerMeeting)
	r.GET("/filesMeeting/:id", controllers.GetFilesByIDMeeting)

	// ********** Route Request ********** //
	r.GET("/request", controllers.RequestIndex)
	r.GET("/AccRequest/:id", controllers.AccRequest)
	r.GET("/CancelRequest/:id", controllers.CancelRequest)

	// ********** Route Notification ********** //
	r.GET("/notifications", utils.GetNotifications)
	r.DELETE("/notifications/:id", utils.DeleteNotification)

	r.Run(":8083")
}
