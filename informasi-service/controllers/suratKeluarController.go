package controllers

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/arkaramadhan/its-vo/common/initializers"
	helper "github.com/arkaramadhan/its-vo/common/utils"
	"github.com/arkaramadhan/its-vo/informasi-service/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type SuratKeluarRequest struct {
	ID       uint    `gorm:"primaryKey"`
	NoSurat  *string `json:"no_surat"`
	Title    *string `json:"title"`
	From     *string `json:"from"`
	Pic      *string `json:"pic"`
	Tanggal  *string `json:"tanggal"`
	CreateBy string  `json:"create_by"`
}

func UploadHandlerSuratKeluar(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/suratkeluar")
}

func GetFilesByIDSuratKeluar(c *gin.Context) {
	helper.GetFilesByID(c)
}

func DeleteFileHandlerSuratKeluar(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/suratkeluar")
}

func DownloadFileHandlerSuratKeluar(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/suratkeluar")
}

func SuratKeluarCreate(c *gin.Context) {
	// Get data off req body
	var requestBody SuratKeluarRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.Status(400)
		c.Error(err) // log the error
		return
	}

	// Add some logging to see what's being received
	log.Println("Received request body:", requestBody)

	requestBody.CreateBy = c.MustGet("username").(string)

	var tanggal *time.Time // Deklarasi variabel tanggal sebagai pointer ke time.Time
	if requestBody.Tanggal != nil && *requestBody.Tanggal != "" {
		// Parse the date string only if it's not nil and not empty
		parsedTanggal, err := time.Parse("2006-01-02", *requestBody.Tanggal)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			c.JSON(400, gin.H{"message": "Invalid format tanggal: " + err.Error()})
			return
		}
		tanggal = &parsedTanggal
	}

	surat_keluar := models.SuratKeluar{
		NoSurat:  requestBody.NoSurat,
		Title:    requestBody.Title,
		From:     requestBody.From,
		Pic:      requestBody.Pic,
		Tanggal:  tanggal, // Gunakan tanggal yang telah diparsing, bisa jadi nil jika input kosong
		CreateBy: requestBody.CreateBy,
	}

	if err := initializers.DB.Create(&surat_keluar).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal membuat surat keluar: " + err.Error()})
		return
	}

	// Return it
	c.JSON(200, gin.H{"message": "surat keluar berhasil dibuat"})
}

func SuratKeluarIndex(c *gin.Context) {
	var surat_keluars []models.SuratKeluar
	helper.FetchAllRecords(initializers.DB, c, &surat_keluars, "informasi.surat_keluars", "Gagal mengambil data surat keluar")
}

func SuratKeluarShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var bc models.SuratKeluar
	helper.ShowRecord(c, initializers.DB, id, &bc, "surat keluar berhasil dilihat", "informasi.surat_keluars")
}

func SuratKeluarUpdate(c *gin.Context) {
	var requestBody SuratKeluarRequest
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
	var bc models.SuratKeluar
	if err := initializers.DB.First(&bc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "surat keluar tidak ditemukan"})
        return
    }

    // Update the memo with new data
	if requestBody.NoSurat != nil {
		bc.NoSurat = requestBody.NoSurat
	}
	if requestBody.Title != nil {
		bc.Title = requestBody.Title
	}
	if requestBody.From != nil {
		bc.From = requestBody.From
	}
	if requestBody.Pic != nil {
		bc.Pic = requestBody.Pic
	}
	if tanggal != nil {
		bc.Tanggal = tanggal
	}
	
	if err := initializers.DB.Save(&bc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update surat keluar: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "berita acara berhasil diupdate"})
}

func SuratKeluarDelete(c *gin.Context) {
	var bc models.SuratKeluar
	helper.DeleteRecordByID(c, initializers.DB, "informasi.surat_keluars", &bc, "surat keluar")
}

func ExportSuratKeluarHandler(c *gin.Context) {
	f := excelize.NewFile()
	ExportSuratKeluarToExcel(c, f, "SURAT KELUAR", true)
}

func ExportSuratKeluarToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var surat_keluar []models.SuratKeluar
	initializers.DB.Find(&surat_keluar)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, sk := range surat_keluar {
		excelData = append(excelData, &sk)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "SURAT KELUAR",
		Columns: []helper.ExcelColumn{
			{Header: "Tanggal", Width: 20},
			{Header: "No Surat", Width: 27},
			{Header: "Title", Width: 40},
			{Header: "From", Width: 20},
			{Header: "PIC", Width: 20},
		},
		Data:         excelData,
		IsSplitSheet: false,
	}

	if f != nil {
		helper.ExportToSheet(f, config)
	} else {
		helper.ExportToExcel(config)
	}

	if isStandAlone {
		fileName := "its_report_suratkeluar.xlsx"
		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Type", "application/octet-stream")
		if err := f.Write(c.Writer); err != nil {
			return err
		}
	}

	return nil
}

func excelDateToTimeSuratKeluar(excelDate int) (time.Time, error) {
	baseDate := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	days := time.Duration(excelDate) * 24 * time.Hour
	return baseDate.Add(days), nil
}

func ImportExcelSuratKeluar(c *gin.Context) {
	// Mengambil file dari form upload
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error retrieving the file: %v", err)
		return
	}
	defer file.Close()

	// Simpan file sementara jika perlu
	tempFile, err := os.CreateTemp("", "*.xlsx")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating temporary file: %v", err)
		return
	}
	defer os.Remove(tempFile.Name()) // Hapus file sementara setelah selesai

	// Salin file dari request ke file sementara
	if _, err := file.Seek(0, 0); err != nil {
		c.String(http.StatusInternalServerError, "Error seeking file: %v", err)
		return
	}
	if _, err := io.Copy(tempFile, file); err != nil {
		c.String(http.StatusInternalServerError, "Error copying file: %v", err)
		return
	}

	// Buka file Excel dari file sementara
	tempFile.Seek(0, 0) // Reset pointer ke awal file
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file: %v", err)
		return
	}
	defer f.Close()

	// Pilih sheet
	sheetName := "SURAT KELUAR"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting rows: %v", err)
		return
	}

	log.Println("Processing rows...") // Log untuk memulai proses baris

	// Definisikan semua format tanggal yang mungkin
	dateFormats := []string{
		"2 January 2006",
		"02-Jan-06",
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
		"06-Jan-02",
		"02-Jan-06",
		"1-Jan-06",
		"06-Jan-02",
	}

	// Loop melalui baris dan simpan ke database
	for i, row := range rows {
		if i == 0 { // Lewati baris pertama yang merupakan header
			continue
		}
		if len(row) < 5 { // Pastikan ada cukup kolom
			log.Printf("Row %d skipped: less than 5 columns filled", i+1)
			continue
		}
		noSurat := row[0]
		title := row[1]
		from := row[2]
		pic := row[3]
		tanggalString := row[4]

		var tanggal time.Time
		var parseErr error

		// Coba konversi dari serial Excel jika tanggalString adalah angka
		if serial, err := strconv.Atoi(tanggalString); err == nil {
			tanggal, parseErr = excelDateToTimeSuratKeluar(serial)
		} else {
			// Coba parse menggunakan format tanggal yang sudah ada
			for _, format := range dateFormats {
				tanggal, parseErr = time.Parse(format, tanggalString)
				if parseErr == nil {
					break // Keluar dari loop jika parsing berhasil
				}
			}
		}

		if parseErr != nil {
			log.Printf("Format tanggal tidak valid di baris %d: %v", i+1, parseErr)
			continue // Lewati baris ini jika format tanggal tidak valid
		}

		// Buat instance baru dari models.SuratKeluar untuk setiap iterasi loop
		surat_keluar := models.SuratKeluar{
			NoSurat:  &noSurat,
			Title:    &title,
			From:     &from,
			Pic:      &pic,
			Tanggal:  &tanggal,
			CreateBy: c.MustGet("username").(string),
		}

		// Simpan ke database
		if err := initializers.DB.Create(&surat_keluar).Error; err != nil {
			log.Printf("Error saving record from row %d: %v", i+1, err)
			c.String(http.StatusInternalServerError, "Error saving record from row %d: %v", i+1, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimpor."})
}
