package main

import (
	"log"
	"os"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/middleware"
	exportAll "github.com/arkaramadhan/its-vo/common/exportAll"
	"github.com/arkaramadhan/its-vo/project-service/controllers"
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

	err := initializers.ConnectToDB("project")
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

	// ********** Project ********** //
	r.POST("/Project", controllers.ProjectCreate)
	r.PUT("/Project/:id", controllers.ProjectUpdate)
	r.GET("/Project", controllers.ProjectIndex)
	r.GET("/Project/:id", controllers.ProjectShow)
	r.DELETE("/Project/:id", controllers.ProjectDelete)
	r.GET("/exportProject", controllers.ExportProjectHandler)
	r.POST("/uploadProject", controllers.ImportExcelProject)

	r.POST("/uploadFileProject", controllers.UploadHandlerProject)
	r.GET("/downloadProject/:id/:filename", controllers.DownloadFileHandlerProject)
	r.DELETE("/deleteProject/:id/:filename", controllers.DeleteFileHandlerProject)
	r.GET("/filesProject/:id", controllers.GetFilesByIDProject)

	r.GET("/exportAll", exportAll.ExportAll)

	r.Run(":8086")
}
