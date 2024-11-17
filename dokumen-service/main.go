package main

import (
	"log"
	"os"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/middleware"
	"github.com/arkaramadhan/its-vo/dokumen-service/controllers"
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

	err := initializers.ConnectToDB("dokumen")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func main() {

	r := gin.Default()

	r.Use(middleware.CORS())

	// ********** Middleware ********** //
	r.Use(middleware.TokenAuthMiddleware())
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// ********** Routes for Memo ********** //
	r.GET("/memo", controllers.MemoIndex)
	r.POST("/memo", controllers.MemoCreate)
	r.GET("/memo/:id", controllers.MemoShow)
	r.PUT("/memo/:id", controllers.MemoUpdate)
	r.DELETE("/memo/:id", controllers.MemoDelete)
	r.GET("/exportMemo", controllers.ExportMemoHandler)
	r.POST("/uploadMemo", controllers.ImportExcelMemo)

	r.POST("/uploadFileMemo", controllers.UploadHandlerMemo)
	r.GET("/downloadMemo/:id/:filename", controllers.DownloadFileHandlerMemo)
	r.DELETE("/deleteMemo/:id/:filename", controllers.DeleteFileHandlerMemo)
	r.GET("/filesMemo/:id", controllers.GetFilesByIDMemo)

	// ********** Routes for Berita Acara ********** //
	r.GET("/beritaAcara", controllers.BeritaAcaraIndex)
	r.POST("/beritaAcara", controllers.BeritaAcaraCreate)
	r.GET("/beritaAcara/:id", controllers.BeritaAcaraShow)
	r.PUT("/beritaAcara/:id", controllers.BeritaAcaraUpdate)
	r.DELETE("/beritaAcara/:id", controllers.BeritaAcaraDelete)
	r.GET("/exportBeritaAcara", controllers.ExportBeritaAcaraHandler)
	r.POST("/uploadBeritaAcara", controllers.ImportExcelBeritaAcara)

	r.POST("/uploadFileBeritaAcara", controllers.UploadHandlerBeritaAcara)
	r.GET("/downloadBeritaAcara/:id/:filename", controllers.DownloadFileHandlerBeritaAcara)
	r.DELETE("/deleteBeritaAcara/:id/:filename", controllers.DeleteFileHandlerBeritaAcara)
	r.GET("/filesBeritaAcara/:id", controllers.GetFilesByIDBeritaAcara)

	// ********** Routes for Surat ********** //
	r.GET("/surat", controllers.SuratIndex)
	r.POST("/surat", controllers.SuratCreate)
	r.GET("/surat/:id", controllers.SuratShow)
	r.PUT("/surat/:id", controllers.SuratUpdate)
	r.DELETE("/surat/:id", controllers.SuratDelete)
	r.GET("/exportSurat", controllers.ExportSuratHandler)
	r.POST("/uploadSurat", controllers.ImportExcelSurat)

	r.POST("/uploadFileSurat", controllers.UploadHandlerSurat)
	r.GET("/downloadSurat/:id/:filename", controllers.DownloadFileHandlerSurat)
	r.DELETE("/deleteSurat/:id/:filename", controllers.DeleteFileHandlerSurat)
	r.GET("/filesSurat/:id", controllers.GetFilesByIDSurat)

	// ********** Routes for SK ********** //
	r.GET("/sk", controllers.SkIndex)
	r.POST("/sk", controllers.SkCreate)
	r.GET("/sk/:id", controllers.SkShow)
	r.PUT("/sk/:id", controllers.SkUpdate)
	r.DELETE("/sk/:id", controllers.SkDelete)
	r.GET("/exportSk", controllers.ExportSkHandler)
	r.POST("/uploadSk", controllers.ImportExcelSk)

	r.POST("/uploadFileSk", controllers.UploadHandlerSk)
	r.GET("/downloadSk/:id/:filename", controllers.DownloadFileHandlerSk)
	r.DELETE("/deleteSk/:id/:filename", controllers.DeleteFileHandlerSk)
	r.GET("/filesSk/:id", controllers.GetFilesByIDSk)

	// ********** Routes for Perdin ********** //
	r.POST("/Perdin", controllers.PerdinCreate)
	r.PUT("/Perdin/:id", controllers.PerdinUpdate)
	r.GET("/Perdin", controllers.PerdinIndex)
	r.DELETE("/Perdin/:id", controllers.PerdinDelete)
	r.GET("/Perdin/:id", controllers.PerdinShow)
	r.GET("/exportPerdin", controllers.ExportPerdinHandler)
	r.POST("/uploadPerdin", controllers.ImportExcelPerdin)

	r.POST("/uploadFilePerdin", controllers.UploadHandlerPerdin)
	r.GET("/downloadPerdin/:id/:filename", controllers.DownloadFileHandlerPerdin)
	r.DELETE("/deletePerdin/:id/:filename", controllers.DeleteFileHandlerPerdin)
	r.GET("/filesPerdin/:id", controllers.GetFilesByIDPerdin)

	r.Run(":8081")
}
