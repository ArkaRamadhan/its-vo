package main

import (
	"log"
	"os"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/middleware"
	exportAll "github.com/arkaramadhan/its-vo/common/exportAll"
	utils "github.com/arkaramadhan/its-vo/common/utils"
	"github.com/arkaramadhan/its-vo/weeklyTimeline-service/controllers"
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

	err := initializers.ConnectToDB("weekly_timeline")
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

	// ********** Route Meeting Schedule ********** //
	r.GET("/meetingSchedule", controllers.MeetingListIndex)
	r.POST("/meetingSchedule", controllers.MeetingListCreate)
	r.GET("/meetingSchedule/:id", controllers.MeetingListShow)
	r.PUT("/meetingSchedule/:id", controllers.MeetingListUpdate)
	r.DELETE("/meetingSchedule/:id", controllers.MeetingListDelete)
	r.GET("/exportMeetingList", controllers.CreateExcelMeetingList)
	r.POST("/uploadMeetingList", controllers.ImportExcelMeetingList)

	r.POST("/uploadFileMeetingList", controllers.UploadHandlerMeetingList)
	r.GET("/downloadMeetingList/:id/:filename", controllers.DownloadFileHandlerMeetingList)
	r.DELETE("/deleteMeetingList/:id/:filename", controllers.DeleteFileHandlerMeetingList)
	r.GET("/filesMeetingList/:id", controllers.GetFilesByIDMeetingList)

	// ********** Route Timeline Project ********** //
	r.GET("/timelineProject", controllers.GetEventsProject)
	r.POST("/timelineProject", controllers.CreateEventProject)
	r.DELETE("/timelineProject/:id", controllers.DeleteEventProject)
	r.GET("/resourceProject", controllers.GetResourcesProject)
	r.POST("/resourceProject", controllers.CreateResourceProject)
	r.DELETE("/resourceProject/:id", controllers.DeleteResourceProject)
	r.GET("/exportTimelineProject", controllers.ExportTimelineProjectHandler)

	r.GET("/notifications", utils.GetNotifications)
	r.DELETE("/notifications/:id", utils.DeleteNotification)

	r.GET("/exportAll", exportAll.ExportAll)

	r.Run(":8085")
}
