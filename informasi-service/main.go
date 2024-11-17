package main

import (
	"log"
	"os"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/middleware"
	"github.com/arkaramadhan/its-vo/informasi-service/controllers"
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

	err := initializers.ConnectToDB("informasi")
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

	// *********** Route Surat Masuk *********** //
	r.POST("/SuratMasuk", controllers.SuratMasukCreate)
	r.PUT("/SuratMasuk/:id", controllers.SuratMasukUpdate)
	r.GET("/SuratMasuk", controllers.SuratMasukIndex)
	r.DELETE("/SuratMasuk/:id", controllers.SuratMasukDelete)
	r.GET("/SuratMasuk/:id", controllers.SuratMasukShow)
	r.GET("/exportSuratMasuk", controllers.ExportSuratMasukHandler)
	r.POST("/importSuratMasuk", controllers.ImportExcelSuratMasuk)

	r.POST("/uploadFileSuratMasuk", controllers.UploadHandlerSuratMasuk)
	r.GET("/downloadSuratMasuk/:id/:filename", controllers.DownloadFileHandlerSuratMasuk)
	r.DELETE("/deleteSuratMasuk/:id/:filename", controllers.DeleteFileHandlerSuratMasuk)
	r.GET("/filesSuratMasuk/:id", controllers.GetFilesByIDSuratMasuk)

	// *********** Route Surat Masuk *********** //
	r.POST("/SuratKeluar", controllers.SuratKeluarCreate)
	r.PUT("/SuratKeluar/:id", controllers.SuratKeluarUpdate)
	r.GET("/SuratKeluar", controllers.SuratKeluarIndex)
	r.DELETE("/SuratKeluar/:id", controllers.SuratKeluarDelete)
	r.GET("/SuratKeluar/:id", controllers.SuratKeluarShow)
	r.GET("/exportSuratKeluar", controllers.ExportSuratKeluarHandler)
	r.POST("/importSuratKeluar", controllers.ImportExcelSuratKeluar)

	r.POST("/uploadFileSuratKeluar", controllers.UploadHandlerSuratKeluar)
	r.GET("/downloadSuratKeluar/:id/:filename", controllers.DownloadFileHandlerSuratKeluar)
	r.DELETE("/deleteSuratKeluar/:id/:filename", controllers.DeleteFileHandlerSuratKeluar)
	r.GET("/filesSuratKeluar/:id", controllers.GetFilesByIDSuratKeluar)

	// *********** Route Arsip *********** //
	r.GET("/Arsip", controllers.ArsipIndex)
	r.POST("/Arsip", controllers.ArsipCreate)
	r.PUT("/Arsip/:id", controllers.ArsipUpdate)
	r.GET("/Arsip/:id", controllers.ArsipShow)
	r.DELETE("/Arsip/:id", controllers.ArsipDelete)
	r.GET("/exportArsip", controllers.ExportArsipHandler)
	r.POST("/uploadArsip", controllers.ImportExcelArsip)

	r.POST("/uploadFileArsip", controllers.UploadHandlerArsip)
	r.GET("/filesArsip/:id", controllers.GetFilesByIDArsip)
	r.GET("/downloadArsip/:id/:filename", controllers.DownloadFileHandlerArsip)
	r.DELETE("/deleteArsip/:id/:filename", controllers.DeleteFileHandlerArsip)

	r.Run(":8082")
}
