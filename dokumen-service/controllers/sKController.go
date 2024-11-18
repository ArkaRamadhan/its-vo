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
	helper.GetFilesByID(c)
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
    initializers.DB.Find(&sKs)

    // 2. Konversi ke interface ExcelData
    var excelData []helper.ExcelData
    for _, sk := range sKs {
        excelData = append(excelData, &sk)
    }

    // 3. Siapkan konfigurasi
    config := helper.ExcelConfig{
        SheetName: "PERDIN",
        Columns: []helper.ExcelColumn{
            {Header: "Tanggal", Width: 20},
            {Header: "No Surat", Width: 27},
            {Header: "Perihal", Width: 40},
            {Header: "PIC", Width: 20},
        },
        Data:         excelData,
        IsSplitSheet: true,
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

func ImportExcelSk(c *gin.Context) {
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

	sheetName := "SK"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting rows: %v", err)
		return
	}

	log.Println("Processing rows...")

	// Definisikan semua format tanggal yang mungkin
	dateFormats := []string{
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
		"06-Jan-02",
		"02-Jan-06",
		"1-Jan-06",
		"06-Jan-02",
	}

	for i, row := range rows {
		if i == 1 { // Lewati baris pertama yang merupakan header
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

		// Ambil data dari kolom SAG (kiri) dengan penanganan jika kolom kosong
		tanggalSAGStr := ""
		if len(row) > 0 {
			tanggalSAGStr = row[0]
		}
		noSuratSAG := ""
		if len(row) > 1 {
			noSuratSAG = row[1]
		}
		perihalSAG := ""
		if len(row) > 2 {
			perihalSAG = row[2]
		}
		picSAG := ""
		if len(row) > 3 {
			picSAG = row[3]
		}

		var tanggalSAG *time.Time
		var parseErr error
		if tanggalSAGStr != "" {
			// Coba parse menggunakan format tanggal yang sudah ada
			for _, format := range dateFormats {
				var parsedTanggal time.Time
				parsedTanggal, parseErr = time.Parse(format, tanggalSAGStr)
				if parseErr == nil {
					tanggalSAG = &parsedTanggal
					break // Keluar dari loop jika parsing berhasil
				}
			}
			if parseErr != nil {
				log.Printf("Format tanggal tidak valid di baris %d: %v", i+1, parseErr)
			}
		}

		skSAG := models.Sk{
			Tanggal:  tanggalSAG,
			NoSurat:  &noSuratSAG,
			Perihal:  &perihalSAG,
			Pic:      &picSAG,
			CreateBy: c.MustGet("username").(string),
		}

		if err := initializers.DB.Create(&skSAG).Error; err != nil {
			log.Printf("Error saving SAG record from row %d: %v", i+1, err)
		} else {
			log.Printf("SAG Row %d imported successfully", i+1)
		}
	}

	// // Proses data ISO
	// for i, row := range rows {
	// 	if i == 0 {
	// 		continue
	// 	}
	// 	if len(row) < 8 { // Pastikan ada cukup kolom untuk ISO
	// 		log.Printf("Row %d skipped: less than 8 columns filled", i+1)
	// 		continue
	// 	}

	// 	// Ambil data dari kolom ISO (kanan)
	// 	tanggalISOStr := row[5]
	// 	noSuratISO := row[6]
	// 	perihalISO := row[7]
	// 	picISO := row[8]

	// 	var tanggalISO time.Time
	// 	var parseErr error

	// 	// Coba konversi dari serial Excel jika tanggalISOStr adalah angka
	// 	if serial, err := strconv.Atoi(tanggalISOStr); err == nil {
	// 		tanggalISO, parseErr = excelDateToTimeMemo(serial)
	// 	} else {
	// 		// Coba parse menggunakan format tanggal yang sudah ada
	// 		for _, format := range dateFormats {
	// 			tanggalISO, parseErr = time.Parse(format, tanggalISOStr)
	// 			if parseErr == nil {
	// 				break // Keluar dari loop jika parsing berhasil
	// 			}
	// 		}
	// 	}

	// 	if parseErr != nil {
	// 		log.Printf("Format tanggal tidak valid di baris %d: %v", i+1, parseErr)
	// 		continue // Lewati baris ini jika format tanggal tidak valid
	// 	}

	// 	skISO := models.Sk{
	// 		Tanggal:  &tanggalISO,
	// 		NoSurat:  &noSuratISO,
	// 		Perihal:  &perihalISO,
	// 		Pic:      &picISO,
	// 		CreateBy: c.MustGet("username").(string),
	// 	}

	// 	if err := initializers.DB.Create(&skISO).Error; err != nil {
	// 		log.Printf("Error saving ISO record from row %d: %v", i+1, err)
	// 	} else {
	// 		log.Printf("ISO Row %d imported successfully", i+1)
	// 	}
	// }

	c.JSON(http.StatusOK, gin.H{"message": "data berhasil diimport"})
}
