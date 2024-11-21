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

type SuratRequest struct {
	ID       uint    `gorm:"primaryKey"`
	Tanggal  *string `json:"tanggal"`
	NoSurat  *string `json:"no_surat"`
	Perihal  *string `json:"perihal"`
	Pic      *string `json:"pic"`
	CreateBy string  `json:"create_by"`
}

func UploadHandlerSurat(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/surat")
}

func GetFilesByIDSurat(c *gin.Context) {
	helper.GetFilesByID(c, "/app/UploadedFile/surat")
}

func DeleteFileHandlerSurat(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/surat") 
}

func DownloadFileHandlerSurat(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/surat")
}

func GetLatestSuratNumber(category string) (string, error) {
	var lastSurat models.Surat
	if category != "ITS-SAG" && category != "ITS-ISO" {
		return "", fmt.Errorf("kategori tidak valid")
	}
	return helper.GetLatestDocumentNumber(strings.TrimPrefix(category, "ITS-"), "S", &lastSurat, "no_surat", "NoSurat", "dokumen.surats")
}

func SuratIndex(c *gin.Context) {
	var surats []models.Surat
	helper.FetchAllRecords(initializers.DB, c, &surats, "dokumen.surats", "Gagal mengambil data surat")
}

func SuratCreate(c *gin.Context) {
	var requestBody SuratRequest

	if err := c.BindJSON(&requestBody); err != nil {
		helper.RespondError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	log.Println("Received request body:", requestBody)

	var tanggal *time.Time
	if requestBody.Tanggal != nil && *requestBody.Tanggal != "" {
		parsedTanggal, err := time.Parse("2006-01-02", *requestBody.Tanggal)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			helper.RespondError(c, http.StatusBadRequest, "Invalid format tanggal: "+err.Error())
			return
		}
		tanggal = &parsedTanggal
	}

	log.Printf("Parsed date: %v", tanggal) // Tambahkan log ini untuk melihat tanggal yang diparsing

	kategori := *requestBody.NoSurat
	nomor, err := GetLatestSuratNumber(kategori)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Failed to get latest surat number")
		return
	}

	requestBody.NoSurat = &nomor

	requestBody.CreateBy = c.MustGet("username").(string)

	surat := models.Surat{
		Tanggal:  tanggal,             // Gunakan tanggal yang telah diparsing, bisa jadi nil jika input kosong
		NoSurat:  requestBody.NoSurat, // Menggunakan NoMemo yang sudah diformat
		Perihal:  requestBody.Perihal,
		Pic:      requestBody.Pic,
		CreateBy: requestBody.CreateBy,
	}

	result := initializers.DB.Create(&surat)
	if result.Error != nil {
		log.Printf("Error saving surat: %v", result.Error)
		helper.RespondError(c, http.StatusInternalServerError, "gagal membuat surat: "+result.Error.Error())
		return
	}
	log.Printf("Surat created successfully: %v", surat)

	c.JSON(http.StatusCreated, gin.H{"message": "surat berhasil dibuat"})
}

func SuratShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var bc models.Surat
	helper.ShowRecord(c, initializers.DB, id, &bc, "surat berhasil dilihat", "dokumen.surats")
}

func SuratUpdate(c *gin.Context) {
	var requestBody SuratRequest
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
	var surat models.Surat
	if err := initializers.DB.First(&surat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "surat tidak ditemukan"})
		return
	}

	// Update the memo with new data
	if tanggal != nil {
		surat.Tanggal = tanggal
	}
	if requestBody.Perihal != nil {
		surat.Perihal = requestBody.Perihal
	}
	if requestBody.Pic != nil {
		surat.Pic = requestBody.Pic
	}

	if err := initializers.DB.Save(&surat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update surat: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Surat updated successfully"})
}

func SuratDelete(c *gin.Context) {
	var bc models.Surat
	helper.DeleteRecordByID(c, initializers.DB, "dokumen.surats", &bc, "surat")
}

func ExportSuratHandler(c *gin.Context) {
	f := excelize.NewFile()
	ExportSuratToExcel(c, f, "SURAT", true)
}

func ExportSuratToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var surats []models.Surat
	initializers.DB.Table("dokumen.surats").Find(&surats)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, surat := range surats {
		excelData = append(excelData, &surat)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "SURAT",
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
		fileName := "its_report_beritaAcara.xlsx"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		if err := f.Write(c.Writer); err != nil {
			return err
		}
	}

	return nil
}

func ImportExcelSurat(c *gin.Context) {
	config := helper.ExcelImportConfig{
		SheetName:   "SURAT",
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
			surat := models.Surat{
				Tanggal:  tanggal,
				NoSurat:  &noSurat,
				Perihal:  &perihal,
				Pic:      &pic,
				CreateBy: c.MustGet("username").(string),
			}

			return initializers.DB.Create(&surat).Error
		},
	}

	if err := helper.ImportExcelFile(c, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
