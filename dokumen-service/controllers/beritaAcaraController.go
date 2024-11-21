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

type BcRequest struct {
	ID       uint    `gorm:"primaryKey"`
	Tanggal  *string `json:"tanggal"`
	NoSurat  *string `json:"no_surat"`
	Perihal  *string `json:"perihal"`
	Pic      *string `json:"pic"`
	CreateBy string  `json:"create_by"`
}

func UploadHandlerBeritaAcara(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/beritaacara")
}

func GetFilesByIDBeritaAcara(c *gin.Context) {
	helper.GetFilesByID(c, "/app/UploadedFile/beritaacara")
}

func DeleteFileHandlerBeritaAcara(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/beritaacara")
}

func DownloadFileHandlerBeritaAcara(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/beritaacara")
}

func BeritaAcaraIndex(c *gin.Context) {
	var beritaAcaras []models.BeritaAcara
	helper.FetchAllRecords(initializers.DB, c, &beritaAcaras, "dokumen.berita_acaras", "Gagal mengambil data berita acara")
}

func GetLatestBeritaAcaraNumber(category string) (string, error) {
	var lastBA models.BeritaAcara
	if category != "ITS-SAG" && category != "ITS-ISO" {
		return "", fmt.Errorf("kategori tidak valid")
	}
	return helper.GetLatestDocumentNumber(strings.TrimPrefix(category, "ITS-"), "BA", &lastBA, "no_surat", "NoSurat", "dokumen.berita_acaras")
}

func BeritaAcaraCreate(c *gin.Context) {
	var requestBody BcRequest

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

	log.Printf("Parsed date: %v", tanggal)

	kategori := *requestBody.NoSurat
	nomor, err := GetLatestBeritaAcaraNumber(kategori)
	if err != nil {
		helper.RespondError(c, http.StatusInternalServerError, "Failed to get latest BA number: "+err.Error())
		return
	}

	// Langsung gunakan `nomor` yang sudah diformat dengan benar
	requestBody.NoSurat = &nomor

	requestBody.CreateBy = c.MustGet("username").(string)

	bc := models.BeritaAcara{
		Tanggal:  tanggal,
		NoSurat:  requestBody.NoSurat,
		Perihal:  requestBody.Perihal,
		Pic:      requestBody.Pic,
		CreateBy: requestBody.CreateBy,
	}

	if err := initializers.DB.Create(&bc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal membuat berita acara: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "berita acara berhasil dibuat"})
}

func BeritaAcaraShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var bc models.BeritaAcara
	helper.ShowRecord(c, initializers.DB, id, &bc, "berita acara berhasil dilihat", "dokumen.berita_acaras")
}

func BeritaAcaraUpdate(c *gin.Context) {
	var requestBody BcRequest
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
	var bc models.BeritaAcara
	if err := initializers.DB.First(&bc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "berita acara tidak ditemukan"})
		return
	}

	// Mengambil nomor surat terbaru
	nomor, err := GetLatestBeritaAcaraNumber(*requestBody.NoSurat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest Berita Acara number"})
		return
	}

	// Update the memo with new data
	if tanggal != nil {
		bc.Tanggal = tanggal
	}
	if requestBody.Perihal != nil {
		bc.Perihal = requestBody.Perihal
	}
	if requestBody.Pic != nil {
		bc.Pic = requestBody.Pic
	}
	if requestBody.NoSurat != nil && *requestBody.NoSurat != "" {
		bc.NoSurat = &nomor
	}

	if err := initializers.DB.Save(&bc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update berita acara: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "berita acara berhasil diupdate"})
}

func BeritaAcaraDelete(c *gin.Context) {
	var bc models.BeritaAcara
	helper.DeleteRecordByID(c, initializers.DB, "dokumen.berita_acaras", &bc, "berita acara")
}

func ExportBeritaAcaraHandler(c *gin.Context) {
	f := excelize.NewFile()

	ExportBeritaAcaraToExcel(c, f, "BERITA ACARA", true)
}

func ExportBeritaAcaraToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var beritaAcaras []models.BeritaAcara
	initializers.DB.Table("dokumen.berita_acaras").Find(&beritaAcaras)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, ba := range beritaAcaras {
		excelData = append(excelData, &ba)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "BERITA ACARA",
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

func ImportExcelBeritaAcara(c *gin.Context) {
	config := helper.ExcelImportConfig{
		SheetName:   "BERITA ACARA",
		MinColumns:  2,
		HeaderRows:  1,
		LogProgress: true,
		ProcessRow: func(row []string, rowIndex int) error {
			// Proses data SAG (kolom kiri)
			if len(row) >= 4 && helper.HasNonEmptyColumns(row[:4], 2) {
				tanggalSAG, _ := helper.ParseDateWithFormats(helper.GetColumn(row, 0))
				noSuratSAG := helper.GetColumn(row, 1)
				perihalSAG := helper.GetColumn(row, 2)
				picSAG := helper.GetColumn(row, 3)

				beritaAcaraSAG := models.BeritaAcara{
					Tanggal:  tanggalSAG,
					NoSurat:  &noSuratSAG,
					Perihal:  &perihalSAG,
					Pic:      &picSAG,
					CreateBy: c.MustGet("username").(string),
				}

				if err := initializers.DB.Create(&beritaAcaraSAG).Error; err != nil {
					log.Printf("Error saving SAG record from row %d: %v", rowIndex+1, err)
				}
			}

			// Proses data ISO (kolom kanan)
			if len(row) >= 9 && helper.HasNonEmptyColumns(row[5:9], 2) {
				tanggalISO, _ := helper.ParseDateWithFormats(helper.GetColumn(row, 5))
				noSuratISO := helper.GetColumn(row, 6)
				perihalISO := helper.GetColumn(row, 7)
				picISO := helper.GetColumn(row, 8)

				beritaAcaraISO := models.BeritaAcara{
					Tanggal:  tanggalISO,
					NoSurat:  &noSuratISO,
					Perihal:  &perihalISO,
					Pic:      &picISO,
					CreateBy: c.MustGet("username").(string),
				}

				if err := initializers.DB.Create(&beritaAcaraISO).Error; err != nil {
					log.Printf("Error saving ISO record from row %d: %v", rowIndex+1, err)
				}
			}

			return nil
		},
	}

	if err := helper.ImportExcelFile(c, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
