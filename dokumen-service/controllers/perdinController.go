package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arkaramadhan/its-vo/common/initializers"
	helper "github.com/arkaramadhan/its-vo/common/utils"
	"github.com/arkaramadhan/its-vo/dokumen-service/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type perdinRequest struct {
	ID        uint    `gorm:"primaryKey"`
	NoPerdin  *string `json:"no_perdin"`
	Tanggal   *string `json:"tanggal"`
	Hotel     *string `json:"hotel"`
	Transport *string `json:"transport"`
	CreateBy  string  `json:"create_by"`
}

func UploadHandlerPerdin(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/perdin")
}

func GetFilesByIDPerdin(c *gin.Context) {
	helper.GetFilesByID(c, "/app/UploadedFile/perdin")
}

func DeleteFileHandlerPerdin(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/perdin")
}

func DownloadFileHandlerPerdin(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/perdin")
}

func GetLatestPerdinNumber(category string) (string, error) {
	var lastPerdin models.Perdin
	if !strings.HasPrefix(category, "PD-ITS") {
		return "", fmt.Errorf("kategori tidak valid")
	}
	return helper.GetLatestDocumentNumber(category, "perdin", &lastPerdin, "no_perdin", "NoPerdin", "dokumen.perdins")
}

func PerdinIndex(c *gin.Context) {
	var perdins []models.Perdin
	helper.FetchAllRecords(initializers.DB, c, &perdins, "dokumen.perdins", "Gagal mengambil data perdin")
}

func PerdinCreate(c *gin.Context) {
	var requestBody perdinRequest
	if err := c.BindJSON(&requestBody); err != nil {
		helper.RespondError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	var tanggal *time.Time
	if requestBody.Tanggal != nil && *requestBody.Tanggal != "" {
		parsedTanggal, err := time.Parse("2006-01-02", *requestBody.Tanggal)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "gagal memparsing tanggal: " + err.Error()})
			return
		}
		tanggal = &parsedTanggal
	}

	kategori := "PD-ITS"
	nomor, err := GetLatestPerdinNumber(kategori)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal mendapatkan nomor perdin terakhir: " + err.Error()})
		return
	}

	requestBody.NoPerdin = &nomor

	requestBody.CreateBy = c.MustGet("username").(string)

	// Create perdin with the requestBody data
	perdin := models.Perdin{
		NoPerdin:  requestBody.NoPerdin,
		Tanggal:   tanggal,
		Hotel:     requestBody.Hotel,
		Transport: requestBody.Transport,
		CreateBy:  requestBody.CreateBy,
	}

	if err := initializers.DB.Create(&perdin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal membuat perdin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "perdin berhasil dibuat"})
}

func PerdinShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var bc models.Perdin
	helper.ShowRecord(c, initializers.DB, id, &bc, "perdin berhasil dilihat", "dokumen.perdins")
}

func PerdinUpdate(c *gin.Context) {
	var requestBody perdinRequest
	if err := c.BindJSON(&requestBody); err != nil {
		log.Printf("Error binding JSON: %v", err)
		helper.RespondError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var tanggal *time.Time
	if requestBody.Tanggal != nil {
		dateFormats := []string{"2006-01-02", "2006-01-02T15:04:05Z07:00", "January 2, 2006", "Jan 2, 2006", "02/01/2006"}
		parsedTanggal, err := helper.ParseFlexibleDate(*requestBody.Tanggal, dateFormats)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid format tanggal: " + err.Error()})
			return
		}
		tanggal = parsedTanggal
	}

	// Assuming you are updating a Memo record
	id := c.Param("id") // or however you get the ID
	var perdin models.Perdin
	if err := initializers.DB.First(&perdin, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "perdin tidak ditemukan"})
		return
	}

	// Mengambil nomor surat terbaru
	nomor, err := GetLatestPerdinNumber(*requestBody.NoPerdin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest Perdin number"})
		return
	}

	// Update the memo with new data
	if tanggal != nil {
		perdin.Tanggal = tanggal
	}
	if requestBody.Hotel != nil {
		perdin.Hotel = requestBody.Hotel
	}
	if requestBody.Transport != nil {
		perdin.Transport = requestBody.Transport
	}
	if requestBody.NoPerdin != nil && *requestBody.NoPerdin != "" {
		perdin.NoPerdin = &nomor
	}

	if err := initializers.DB.Save(&perdin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update perdin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "perdin berhasil diupdate"})
}

func PerdinDelete(c *gin.Context) {
	var perdin models.Perdin
	helper.DeleteRecordByID(c, initializers.DB, "dokumen.perdins", &perdin, "perdin")
}

func ExportPerdinHandler(c *gin.Context) {
	f := excelize.NewFile()
	ExportPerdinToExcel(c, f, "PERDIN", true)
}

func ExportPerdinToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var perdins []models.Perdin
	initializers.DB.Table("dokumen.perdins").Find(&perdins)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, ba := range perdins {
		excelData = append(excelData, &ba)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "PERDIN",
		Columns: []helper.ExcelColumn{
			{Header: "Tanggal", Width: 20},
			{Header: "No Perdin", Width: 27},
			{Header: "Hotel", Width: 40},
			{Header: "Transport", Width: 20},
		},
		Data:         excelData,
		IsSplitSheet: true,
		GetStatus:    nil,
		SplitType:    helper.SplitVertical,
	}

	if f != nil {
		helper.ExportToSheet(f, config)
	} else {
		helper.ExportToExcel(config)
	}

	if isStandAlone {
		fileName := "its_report_beritaAcara.xlsx"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		if err := f.Write(c.Writer); err != nil {
			return err
		}
	}

	return nil

}

func ImportExcelPerdin(c *gin.Context) {
	config := helper.ExcelImportConfig{
		SheetName:   "PERDIN",
		MinColumns:  2,
		HeaderRows:  1,
		LogProgress: true,
		ProcessRow: func(row []string, rowIndex int) error {
			// Ambil data dari kolom
			noPerdin := helper.GetColumn(row, 0)
			tanggalStr := helper.GetColumn(row, 1)
			hotel := helper.GetColumn(row, 2)
			transport := helper.GetColumn(row, 3)

			// Parse tanggal
			tanggal, _ := helper.ParseDateWithFormats(tanggalStr)

			// Buat dan simpan memo
			perdin := models.Perdin{
				Tanggal:   tanggal,
				NoPerdin:  &noPerdin,
				Hotel:     &hotel,
				Transport: &transport,
				CreateBy:  c.MustGet("username").(string),
			}

			return initializers.DB.Create(&perdin).Error
		},
	}

	if err := helper.ImportExcelFile(c, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
