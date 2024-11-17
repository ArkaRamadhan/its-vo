package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
	helper.GetFilesByID(c)
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
	if requestBody.NoPerdin != nil {
		perdin.NoPerdin = requestBody.NoPerdin
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
	// 1. Ambil data dari database
	var perdins []models.Perdin
	initializers.DB.Find(&perdins)

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
	}

	// 4. Panggil fungsi ExportToExcel
	f, err := helper.ExportToExcel(config)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengekspor data ke Excel: "+err.Error())
		return
	}

	// 5. Set header dan kirim file
	fileName := "its_report_perdin.xlsx"
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	f.Write(c.Writer)
}

func excelDateToTimePerdin(excelDate int) (time.Time, error) {
	// Excel menggunakan tanggal mulai 1 Januari 1900 (serial 1)
	baseDate := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	days := time.Duration(excelDate) * 24 * time.Hour
	return baseDate.Add(days), nil
}

func ImportExcelPerdin(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error retrieving the file: %v", err)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "*.xlsx")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating temporary file: %v", err)
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := io.Copy(tempFile, file); err != nil {
		c.String(http.StatusInternalServerError, "Error copying file: %v", err)
		return
	}

	tempFile.Seek(0, 0)
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file: %v", err)
		return
	}
	defer f.Close()

	sheetName := "PERDIN"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting rows: %v", err)
		return
	}

	log.Println("Processing rows...")

	// Definisikan semua format tanggal yang mungkin
	dateFormats := []string{
		"02-Jan-06",
		"06-Jan-02",
		"2 January 2006",
		"2006-01-02",
		"02-01-2006",
		"01/02/2006",
		"2006.01.02",
		"02/01/2006",
		"Jan 2, 06",
		"Jan 2, 2006",
		"01/02/06",
		"02/01/06",
		"06/02/01",
		"06/01/02",
		"1-Jan-06",
		"06-Jan-02",
	}

	for i, row := range rows {
		if i == 0 { // Lewati baris pertama yang merupakan header
			continue
		}
		if len(row) < 2 { // Pastikan ada cukup kolom
			log.Printf("Baris %d dilewati: kurang dari 2 kolom terisi", i+1)
			continue
		}

		noPerdin := helper.GetStringOrNil(helper.GetColumn(row, 0))
		tanggalStr := helper.GetStringOrNil(helper.GetColumn(row, 1))
		hotel := helper.GetStringOrNil(helper.GetColumn(row, 2))
		transport := helper.GetStringOrNil(helper.GetColumn(row, 3))

		var tanggalTime *time.Time
		var parseErr error // Deklarasi ulang di setiap iterasi untuk menghindari nilai residual

		if tanggalStr != nil {
			if serial, err := strconv.Atoi(*tanggalStr); err == nil {
				parsed, err := excelDateToTimePerdin(serial)
				if err == nil {
					tanggalTime = &parsed
				} else {
					parseErr = err
				}
			} else {
				for _, format := range dateFormats {
					parsed, err := time.Parse(format, *tanggalStr)
					if err == nil {
						tanggalTime = &parsed
						break
					} else {
						parseErr = err
					}
				}
			}

			if parseErr != nil {
				log.Printf("Format tanggal tidak valid di baris %d: %v", i+1, parseErr)
				continue // Lewati baris ini jika format tanggal tidak valid
			}
		}

		// Buat objek Perdin dengan data yang diambil
		perdin := models.Perdin{
			Tanggal:   tanggalTime,
			NoPerdin:  noPerdin,
			Hotel:     hotel,
			Transport: transport,
			CreateBy:  c.MustGet("username").(string),
		}

		// Simpan objek Perdin ke dalam database
		if err := initializers.DB.Create(&perdin).Error; err != nil {
			log.Printf("Error menyimpan record dari baris %d: %v", i+1, err)
			continue
		}
		log.Printf("Baris %d diimpor dengan sukses", i+1)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
