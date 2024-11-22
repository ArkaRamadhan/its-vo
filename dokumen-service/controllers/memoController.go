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

type MemoRequest struct {
	ID       uint    `gorm:"primaryKey"`
	Tanggal  *string `json:"tanggal"`
	NoMemo   *string `json:"no_memo"`
	Perihal  *string `json:"perihal"`
	Pic      *string `json:"pic"`
	Kategori *string `json:"kategori"`
	CreateBy string  `json:"create_by"`
}

func UploadHandlerMemo(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/memo")
}

func GetFilesByIDMemo(c *gin.Context) {
	helper.GetFilesByID(c, "/app/UploadedFile/memo")
}

func DeleteFileHandlerMemo(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/memo")
}

func DownloadFileHandlerMemo(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/memo")
}

func GetLatestMemoNumber(category string) (string, error) {
	var lastMemo models.Memo
	if category != "ITS-SAG" && category != "ITS-ISO" {
		return "", fmt.Errorf("kategori tidak valid")
	}
	return helper.GetLatestDocumentNumber(strings.TrimPrefix(category, "ITS-"), "M", &lastMemo, "no_memo", "NoMemo", "dokumen.memos")
}

func MemoIndex(c *gin.Context) {
	var memos []models.Memo
	helper.FetchAllRecords(initializers.DB, c, &memos, "dokumen.memos", "Gagal mengambil data memo")
}

func MemoCreate(c *gin.Context) {
	var requestBody MemoRequest

	if err := c.BindJSON(&requestBody); err != nil {
		helper.RespondError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	log.Println("Received request body:", requestBody)

	var tanggal *time.Time
	if requestBody.Tanggal != nil && *requestBody.Tanggal != "" {
		// Coba beberapa format tanggal yang mungkin
		dateFormats := []string{"2006-01-02", "2006-01-02T15:04:05Z07:00", "January 2, 2006", "Jan 2, 2006", "02/01/2006"}
		var parsedTanggal time.Time
		var err error
		for _, format := range dateFormats {
			parsedTanggal, err = time.Parse(format, *requestBody.Tanggal)
			if err == nil {
				tanggal = &parsedTanggal
				break
			}
		}
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "gagal memparsing tanggal: " + err.Error()})
			return
		}
	}

	kategori := *requestBody.NoMemo
	nomor, err := GetLatestMemoNumber(kategori)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal mendapatkan nomor memo terakhir: " + err.Error()})
		return
	}

	requestBody.NoMemo = &nomor
	log.Printf("Generated NoMemo: %s", *requestBody.NoMemo) // Log nomor memo

	requestBody.CreateBy = c.MustGet("username").(string)

	memosag := models.Memo{
		Tanggal:  tanggal,
		NoMemo:   requestBody.NoMemo, // Menggunakan NoMemo yang sudah diformat
		Perihal:  requestBody.Perihal,
		Pic:      requestBody.Pic,
		CreateBy: requestBody.CreateBy,
	}

	result := initializers.DB.Create(&memosag)
	if result.Error != nil {
		log.Printf("Error saving memo: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal membuat memo: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "memo berhasil dibuat"})
}

func MemoShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var memo models.Memo
	helper.ShowRecord(c, initializers.DB, id, &memo, "memo berhasil dilihat", "dokumen.memos")
}

func MemoUpdate(c *gin.Context) {
	var requestBody MemoRequest
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
	var memo models.Memo
	if err := initializers.DB.First(&memo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Memo not found"})
		return
	}

	// Mengambil nomor surat terbaru
	nomor, err := GetLatestMemoNumber(*requestBody.NoMemo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest Memo number"})
		return
	}

	// Update the memo with new data
	if tanggal != nil {
		memo.Tanggal = tanggal
	}
	if requestBody.Perihal != nil {
		memo.Perihal = requestBody.Perihal
	}
	if requestBody.Pic != nil {
		memo.Pic = requestBody.Pic
	}
	if requestBody.NoMemo != nil && *requestBody.NoMemo != "" {
		memo.NoMemo = &nomor
	}
	if err := initializers.DB.Save(&memo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update memo: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memo updated successfully"})
}

func MemoDelete(c *gin.Context) {
	var memosag models.Memo
	helper.DeleteRecordByID(c, initializers.DB, "dokumen.memos", &memosag, "memo")
}

func ExportMemoHandler(c *gin.Context) {
	f := excelize.NewFile()
	ExportMemoToExcel(c, f, "MEMO", true)
}

func ExportMemoToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var memos []models.Memo
	initializers.DB.Table("dokumen.memos").Find(&memos)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, ba := range memos {
		excelData = append(excelData, &ba)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "MEMO",
		Columns: []helper.ExcelColumn{
			{Header: "Tanggal", Width: 20},
			{Header: "No Memo", Width: 27},
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
		fileName := "its_report_memo.xlsx"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		if err := f.Write(c.Writer); err != nil {
			return err
		}
	}

	return nil
}

func ImportExcelMemo(c *gin.Context) {
	config := helper.ExcelImportConfig{
		SheetName:   "MEMO",
		MinColumns:  2,
		HeaderRows:  1,
		LogProgress: true,
		ProcessRow: func(row []string, rowIndex int) error {
			// Ambil data dari kolom
			tanggalStr := helper.GetColumn(row, 0)
			noMemo := helper.GetColumn(row, 1)
			perihal := helper.GetColumn(row, 2)
			pic := helper.GetColumn(row, 3)

			// Parse tanggal
			tanggal, _ := helper.ParseDateWithFormats(tanggalStr)

			// Buat dan simpan memo
			memo := models.Memo{
				Tanggal:  tanggal,
				NoMemo:   &noMemo,
				Perihal:  &perihal,
				Pic:      &pic,
				CreateBy: c.MustGet("username").(string),
			}

			return initializers.DB.Create(&memo).Error
		},
	}

	if err := helper.ImportExcelFile(c, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
