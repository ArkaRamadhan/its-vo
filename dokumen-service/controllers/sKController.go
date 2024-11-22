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

type SKRequest struct {
	ID       uint    `gorm:"primaryKey"`
	Tanggal  *string `json:"tanggal"`
	NoSurat  *string `json:"no_surat"`
	Perihal  *string `json:"perihal"`
	Pic      *string `json:"pic"`
	CreateBy string  `json:"create_by"`
}

func UploadHandlerSk(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/sk")
}

func GetFilesByIDSk(c *gin.Context) {
	helper.GetFilesByID(c, "/app/UploadedFile/sk")
}

func DeleteFileHandlerSk(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/sk")
}

func DownloadFileHandlerSk(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/sk")
}

func GetLatestSkNumber(category string) (string, error) {
	var lastSk models.Sk
	if category != "ITS-SAG" && category != "ITS-ISO" {
		return "", fmt.Errorf("kategori tidak valid")
	}
	return helper.GetLatestDocumentNumber(strings.TrimPrefix(category, "ITS-"), "SK", &lastSk, "no_surat", "NoSurat", "dokumen.sks")
}

func SkIndex(c *gin.Context) {
	var sKs []models.Sk
	helper.FetchAllRecords(initializers.DB, c, &sKs, "dokumen.sks", "Gagal mengambil data sk")
}

func SkCreate(c *gin.Context) {
	var requestBody SKRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body: " + err.Error()})
		return
	}

	log.Println("Received request body:", requestBody)

	var tanggal *time.Time
	if requestBody.Tanggal != nil && *requestBody.Tanggal != "" {
		parsedTanggal, err := time.Parse("2006-01-02", *requestBody.Tanggal)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid format tanggal: " + err.Error()})
			return
		}
		tanggal = &parsedTanggal
	}

	log.Printf("Parsed date: %v", tanggal) // Tambahkan log ini untuk melihat tanggal yang diparsing

	nomor, err := GetLatestSkNumber(*requestBody.NoSurat)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Failed to get latest surat number")
		return
	}

	requestBody.NoSurat = &nomor
	requestBody.CreateBy = c.MustGet("username").(string)

	sK := models.Sk{
		Tanggal:  tanggal,             // Gunakan tanggal yang telah diparsing, bisa jadi nil jika input kosong
		NoSurat:  requestBody.NoSurat, // Menggunakan NoMemo yang sudah diformat
		Perihal:  requestBody.Perihal,
		Pic:      requestBody.Pic,
		CreateBy: requestBody.CreateBy,
	}

	result := initializers.DB.Create(&sK)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal membuat surat: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "surat berhasil dibuat"})
}

func SkShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var bc models.Sk
	helper.ShowRecord(c, initializers.DB, id, &bc, "sk berhasil dilihat", "dokumen.sks")
}

func SkUpdate(c *gin.Context) {
	var requestBody SKRequest
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
	var sk models.Sk
	if err := initializers.DB.First(&sk, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "sk tidak ditemukan"})
		return
	}

	// Mengambil nomor surat terbaru
	nomor, err := GetLatestSkNumber(*requestBody.NoSurat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest SK number"})
		return
	}

	// Update the sk with new data
	if tanggal != nil {
		sk.Tanggal = tanggal
	}
	if requestBody.Perihal != nil {
		sk.Perihal = requestBody.Perihal
	}
	if requestBody.Pic != nil {
		sk.Pic = requestBody.Pic
	}
	if requestBody.NoSurat != nil && *requestBody.NoSurat != "" {
		sk.NoSurat = &nomor
	}

	if err := initializers.DB.Save(&sk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update sk: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "surat berhasil diupdate"})
}

func SkDelete(c *gin.Context) {
	var bc models.Sk
	helper.DeleteRecordByID(c, initializers.DB, "dokumen.sks", &bc, "sk")
}

func ExportSkHandler(c *gin.Context) {
	f := excelize.NewFile()
	ExportSkToExcel(c, f, "SK", true)
}

func ExportSkToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var sKs []models.Sk
	initializers.DB.Table("dokumen.sks").Find(&sKs)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, sk := range sKs {
		excelData = append(excelData, &sk)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "SK",
		Columns: []helper.ExcelColumn{
			{Header: "Tanggal", Width: 20},
			{Header: "No Surat", Width: 27},
			{Header: "Perihal", Width: 40},
			{Header: "PIC", Width: 20},
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
		fileName := "its_report_sk.xlsx"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		if err := f.Write(c.Writer); err != nil {
			return err
		}
	}

	return nil
}

func ImportExcelSk(c *gin.Context) {
	config := helper.ExcelImportConfig{
		SheetName:   "SK",
		MinColumns:  2,
		HeaderRows:  1,
		LogProgress: true,
		ProcessRow: func(row []string, rowIndex int) error {
			// Ambil data dari kolom
			tanggalStr := helper.GetColumn(row, 0)
			noSurat := helper.GetColumn(row, 1)
			perihal := helper.GetColumn(row, 2)
			pic := helper.GetColumn(row, 3)
			// Parse tanggal
			tanggal, _ := helper.ParseDateWithFormats(tanggalStr)

			// Buat dan simpan memo
			sk := models.Sk{
				Tanggal:  tanggal,
				NoSurat:  &noSurat,
				Perihal:  &perihal,
				Pic:      &pic,
				CreateBy: c.MustGet("username").(string),
			}

			return initializers.DB.Create(&sk).Error
		},
	}

	if err := helper.ImportExcelFile(c, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
