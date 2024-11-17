package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	helper.GetFilesByID(c)
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
	if requestBody.NoSurat != nil {
		bc.NoSurat = requestBody.NoSurat
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
	// 1. Ambil data dari database
	var beritaAcaras []models.BeritaAcara
	initializers.DB.Find(&beritaAcaras)

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
	}

	// 4. Panggil fungsi ExportToExcel
	f, err := helper.ExportToExcel(config)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengekspor data ke Excel: "+err.Error())
		return
	}

	// 5. Set header dan kirim file
	fileName := "its_report_beritaAcara.xlsx"
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	f.Write(c.Writer)
}

func ImportExcelBeritaAcara(c *gin.Context) {
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

	sheetName := "BERITA ACARA"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting rows: %v", err)
		return
	}

	log.Println("Processing rows...")

	// Definisikan format tanggal untuk SAG
	dateFormatsSAG := []string{
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
	}

	// Definisikan format tanggal untuk ISO
	dateFormatsISO := []string{
		"06-Jan-02",
		"02-Jan-06",
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
	}

	for i, row := range rows {
		if i == 0 { // Lewati baris pertama yang merupakan header
			continue
		}
		// Pastikan minimal 2 kolom terisi
		nonEmptyColumns := 0
		for _, col := range row {
			if col != "" {
				nonEmptyColumns++
			}
		}
		if nonEmptyColumns < 2 {
			log.Printf("Baris %d dilewati: hanya %d kolom terisi", i+1, nonEmptyColumns)
			continue
		}

		// Ambil data SAG dari kolom kiri
		tanggalSAGStr, noSuratSAG, perihalSAG, picSAG := "", "", "", ""
		if len(row) > 0 {
			tanggalSAGStr = row[0]
		}
		if len(row) > 1 {
			noSuratSAG = row[1]
		}
		if len(row) > 2 {
			perihalSAG = row[2]
		}
		if len(row) > 3 {
			picSAG = row[3]
		}

		// Ambil data ISO dari kolom kanan
		tanggalISOStr, noSuratISO, perihalISO, picISO := "", "", "", ""
		if len(row) > 5 {
			tanggalISOStr = row[5]
		}
		if len(row) > 6 {
			noSuratISO = row[6]
		}
		if len(row) > 7 {
			perihalISO = row[7]
		}
		if len(row) > 8 {
			picISO = row[8]
		}

		// Proses tanggal SAG
		var tanggalSAG *time.Time
		if tanggalSAGStr != "" {
			for _, format := range dateFormatsSAG {
				parsedTanggal, err := time.Parse(format, tanggalSAGStr)
				if err == nil {
					tanggalSAG = &parsedTanggal
					break
				}
			}
		}

		// Proses tanggal ISO
		var tanggalISO *time.Time
		if tanggalISOStr != "" {
			for _, format := range dateFormatsISO {
				parsedTanggal, err := time.Parse(format, tanggalISOStr)
				if err == nil {
					tanggalISO = &parsedTanggal
					break
				}
			}
		}

		// Simpan data SAG
		beritaAcaraSAG := models.BeritaAcara{
			Tanggal:  tanggalSAG,
			NoSurat:  &noSuratSAG,
			Perihal:  &perihalSAG,
			Pic:      &picSAG,
			CreateBy: c.MustGet("username").(string),
		}
		if err := initializers.DB.Create(&beritaAcaraSAG).Error; err != nil {
			log.Printf("Error saving SAG record from row %d: %v", i+1, err)
		} else {
			log.Printf("SAG Row %d imported successfully", i+1)
		}

		// Simpan data ISO
		beritaAcaraISO := models.BeritaAcara{
			Tanggal:  tanggalISO,
			NoSurat:  &noSuratISO,
			Perihal:  &perihalISO,
			Pic:      &picISO,
			CreateBy: c.MustGet("username").(string),
		}
		if err := initializers.DB.Create(&beritaAcaraISO).Error; err != nil {
			log.Printf("Error saving ISO record from row %d: %v", i+1, err)
		} else {
			log.Printf("ISO Row %d imported successfully", i+1)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
