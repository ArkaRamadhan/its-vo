package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

)

func GetColumn(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return row[index]
}

// Helper function to return nil if the string is empty
func GetStringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func HasNonEmptyColumns(row []string, minNonEmpty int) bool {
	count := 0
	for _, col := range row {
		if col != "" {
			count++
		}
		if count >= minNonEmpty {
			return true
		}
	}
	return false
}

func CleanNumericString(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, input)
}

func ParseDateImport(dateStr string) (time.Time, error) {
	dateFormats := []string{
		"2 January 2006",
		"02-06",
		"2-January-2006",
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
		"01/06",
		"02/06",
		"Jan-06", // Menambahkan format ini untuk mengenali "Feb-24" sebagai "Feb-2024"
	}

	// Menambahkan logika untuk menangani format "Feb-24"
	if strings.Contains(dateStr, "-") && len(dateStr) == 5 {
		dateStr = dateStr[:3] + "20" + dateStr[4:]
	}

	for _, format := range dateFormats {
		parsedDate, err := time.Parse(format, dateStr)
		if err == nil {
			return parsedDate, nil
		}
	}
	return time.Time{}, fmt.Errorf("no valid date format found")
}

// Helper function to parse date or return nil if input is nil
func ParseDateOrNil(dateStr *string) *time.Time {
	if dateStr == nil {
		return nil
	}
	parsedDate, err := ParseDateImport(*dateStr)
	if err != nil {
		return nil
	}
	return &parsedDate
}

func RespondError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"message": msg})
}

func ParseDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil || *dateStr == "" {
		return nil, nil
	}
	parsedDate, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return nil, fmt.Errorf("Invalid date format: %s", err.Error())
	}
	return &parsedDate, nil
}

// ********** FUNC UPLOAD FILE *********** //

func UploadHandler(c *gin.Context, baseDir string) {
	// mantap min
	id := c.PostForm("id")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "File diperlukan"})
		return
	}

	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID tidak valid"})
		return
	}

	dir := filepath.Join(baseDir, id)

	filePath := filepath.ToSlash(filepath.Join(dir, file.Filename))
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan file"})
		return
	}

	// Simpan metadata ke database
	newFile := models.File{
		UserID:      uint(userID),
		FilePath:    filePath,
		FileName:    file.Filename,
		ContentType: file.Header.Get("Content-Type"),
		Size:        file.Size,
	}
	result := initializers.DB.Table("common.files").Create(&newFile)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan metadata file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File berhasil diunggah"})
}

func GetFilesByID(c *gin.Context, baseDir string) {
	id := c.Param("id")

	filePathPattern := fmt.Sprintf("%s/%s/%%", baseDir, id)

	var files []models.File
	result := initializers.DB.Table("common.files").Where("file_path LIKE ?", filePathPattern).Find(&files)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil data file"})
		return
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.FileName)
	}

	c.JSON(http.StatusOK, fileNames)
}

func DeleteFileHandler(c *gin.Context, mainDir string) {
    encodedFilename := c.Param("filename")
    filename, err := url.QueryUnescape(encodedFilename)
    if err != nil {
        log.Printf("Error decoding filename: %v", err)
        RespondError(c, http.StatusBadRequest, "Invalid filename")
        return
    }

    id := c.Param("id")
    log.Printf("Received ID: %s and Filename: %s", id, filename)

    baseDir := mainDir
    fullPath := filepath.Join(baseDir, id, filename)
    dirPath := filepath.Join(baseDir, id) // Path ke folder ID

    log.Printf("Attempting to delete file at path: %s", fullPath)

    // Hapus file dari sistem file
    err = os.Remove(fullPath)
    if err != nil {
        log.Printf("Error deleting file: %v", err)
        RespondError(c, http.StatusInternalServerError, "Failed to delete file")
        return
    }

    // Hapus metadata file dari database
    result := initializers.DB.Table("common.files").Where("file_path = ?", fullPath).Delete(&models.File{})
    if result.Error != nil {
        log.Printf("Error deleting file metadata: %v", result.Error)
        RespondError(c, http.StatusInternalServerError, "Failed to delete file metadata")
        return
    }

    // Cek apakah folder kosong dan hapus jika kosong
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        log.Printf("Error reading directory %s: %v", dirPath, err)
    } else {
        if len(entries) == 0 {
            log.Printf("Directory is empty, attempting to delete: %s", dirPath)
            if err := os.Remove(dirPath); err != nil {
                log.Printf("Error deleting empty directory %s: %v", dirPath, err)
            } else {
                log.Printf("Successfully deleted empty directory: %s", dirPath)
            }
        } else {
            log.Printf("Directory not empty, contains %d files", len(entries))
            for _, entry := range entries {
                log.Printf("Remaining file: %s", entry.Name())
            }
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "File deleted successfully",
        "path": fullPath,
        "directory_status": map[string]interface{}{
            "path": dirPath,
            "deleted": len(entries) == 0,
        },
    })
}

func DownloadFileHandler(c *gin.Context, mainDir string) {
	id := c.Param("id")
	filename := c.Param("filename")
	baseDir := mainDir
	fullPath := filepath.Join(baseDir, id, filename)

	log.Printf("Full path for download: %s", fullPath)

	// Periksa keberadaan file di database
	var file models.File
	result := initializers.DB.Table("common.files").Where("file_path = ?", fullPath).First(&file)
	if result.Error != nil {
		log.Printf("File not found in database: %v", result.Error)
		RespondError(c, http.StatusNotFound, "File tidak ditemukan")
		return
	}

	// Periksa keberadaan file di sistem file
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Printf("File not found in system: %s", fullPath)
		RespondError(c, http.StatusNotFound, "File tidak ditemukan di sistem file")
		return
	}

	log.Printf("File downloaded successfully: %s", fullPath)
	c.File(fullPath)
}

// ********** FUNC GET LATEST NUMBER ********** //

// GetLatestDocumentNumber menghasilkan nomor dokumen berikutnya berdasarkan kategori dan tipe dokumen
func GetLatestDocumentNumber(category, docType string, model interface{}, dbField, structField string, schema string) (string, error) {
	currentYear := time.Now().Year()
	docType = strings.TrimSpace(strings.ToLower(docType))
	var searchPattern string
	var newNumber string
	if docType == "perdin" {
		searchPattern = fmt.Sprintf("%%/%s/%d", category, currentYear)
	} else {
		searchPattern = fmt.Sprintf("%%/ITS-%s/%s/%d", category, docType, currentYear) // Ini akan mencari format seperti '%/ITS-SAG/M/2023", category, docType)
	}

	result := initializers.DB.Table(schema).
		Where(fmt.Sprintf("%s LIKE ?", dbField), searchPattern).
		Order("id desc").
		First(model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Ini bukan error, ini expected behavior untuk dokumen pertama
			var newNumber string
			docType = strings.TrimSpace(strings.ToLower(docType))
			if docType == "perdin" {
				newNumber = fmt.Sprintf("00001/%s/%d", category, currentYear)
			} else {
				newNumber = fmt.Sprintf("00001/ITS-%s/%s/%d", category, docType, currentYear)
			}
			return newNumber, nil
		}
		// Ini baru error yang sebenarnya
		return "", fmt.Errorf("error saat query database: %v", result.Error)
	}

	val := reflect.ValueOf(model).Elem()
	numberField := val.FieldByName(structField)
	if !numberField.IsValid() || numberField.IsNil() {
		return "", fmt.Errorf("field %s tidak ditemukan atau nil", structField)
	}

	lastNumber := numberField.Interface().(*string)
	if lastNumber == nil {
		return "", fmt.Errorf("nomor dokumen terakhir adalah nil")
	}

	parts := strings.Split(*lastNumber, "/")
	if len(parts) < 1 {
		return "", fmt.Errorf("format nomor tidak valid")
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("gagal mengkonversi nomor: %v", err)
	}

	if docType == "perdin" {
		newNumber = fmt.Sprintf("%05d/%s/%d", num+1, category, currentYear)
	} else {
		newNumber = fmt.Sprintf("%05d/ITS-%s/%s/%d", num+1, category, docType, currentYear)
    }
	
	log.Printf("Berhasil generate nomor baru: %s", newNumber)
	return newNumber, nil
}

// ********** Component CRUD **********//

func FetchAllRecords[T any](db *gorm.DB, c *gin.Context, result *[]T, schema string, failMessage string) {
	if err := db.Table(schema).Find(result).Error; err != nil {
		RespondError(c, http.StatusInternalServerError, failMessage+": "+err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func ShowRecord[T any](c *gin.Context, db *gorm.DB, id string, data *T, successMessage string, schema string) {
	if err := db.Table(schema).First(data, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "record tidak ditemukan"})
		return
	}
	log.Printf("Data retrieved: %+v", data) // Tambahkan log ini untuk melihat data yang diambil
	c.JSON(http.StatusOK, data)
}

// ParseFlexibleDate mencoba mem-parsing string tanggal dengan daftar format yang diberikan.
func ParseFlexibleDate(dateStr string, formats []string) (*time.Time, error) {
	var parsedDate time.Time
	var err error
	for _, format := range formats {
		parsedDate, err = time.Parse(format, dateStr)
		if err == nil {
			return &parsedDate, nil
		}
	}
	return nil, fmt.Errorf("tanggal tidak valid, semua format gagal: %v", err)
}

func DeleteRecordByID(c *gin.Context, db *gorm.DB, schema string, model interface{}, modelName string) {
	id := c.Params.ByName("id")
	var files []models.File
	if err := initializers.DB.Table("common.files").Where("user_id = ?", id).Find(&files).Error; err != nil {
		log.Printf("Error getting files: %v", err)
	}

	// Gunakan UPLOAD_DIR dari environment variable
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		log.Printf("Warning: UPLOAD_DIR tidak diatur, menggunakan path default")
		uploadDir = "/app/UploadedFile"
	}
	
	baseDir := filepath.Join(uploadDir, modelName, id)
	log.Printf("Mencoba menghapus direktori: %s", baseDir)

	if _, err := os.Stat(baseDir); !os.IsNotExist(err) {
		// Hapus file dari database terlebih dahulu
		if err := initializers.DB.Table("common.files").Where("user_id = ?", id).Delete(&models.File{}).Error; err != nil {
			log.Printf("Error saat menghapus record file dari database: %v", err)
		}

		// Hapus file satu per satu
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			log.Printf("Error membaca direktori %s: %v", baseDir, err)
		} else {
			for _, entry := range entries {
				fullPath := filepath.Join(baseDir, entry.Name())
				if err := os.Remove(fullPath); err != nil {
					log.Printf("Error menghapus file %s: %v", fullPath, err)
				} else {
					log.Printf("Berhasil menghapus file: %s", fullPath)
				}
			}
		}

		// Hapus direktori setelah semua file dihapus
		if err := os.RemoveAll(baseDir); err != nil {
			log.Printf("Error saat menghapus direktori %s: %v", baseDir, err)
		} else {
			log.Printf("Berhasil menghapus direktori: %s", baseDir)
		}
	} else {
		log.Printf("Direktori tidak ditemukan atau sudah dihapus: %s", baseDir)
	}

	// Hapus record dari database utama
	if err := db.Table(schema).First(model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": modelName + " tidak ditemukan"})
		return
	}

	if err := db.Table(schema).Delete(model).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "gagal menghapus " + modelName + ": " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": modelName + " berhasil dihapus",
		"details": map[string]interface{}{
			"directory_path": baseDir,
			"model_name": modelName,
			"id": id,
		},
	})
}

// ********** END OF COMPONENT CRUD ********** //

// ********** START OF IMPORT EXCEL ********** //

type DateFormat struct {
	Format      string
	Description string
	Example     string
}

var CommonDateFormats = []DateFormat{
	{Format: "2 January 2006", Description: "Long date", Example: "2 January 2006"},
	{Format: "02-06", Description: "Short month-year", Example: "02-06"},
	{Format: "2-January-2006", Description: "Long date with dash", Example: "2-January-2006"},
	{Format: "2006-01-02", Description: "ISO format", Example: "2006-01-02"},
	{Format: "02-01-2006", Description: "UK format with dash", Example: "02-01-2006"},
	{Format: "01/02/2006", Description: "US format", Example: "01/02/2006"},
	{Format: "2006.01.02", Description: "Dot separated", Example: "2006.01.02"},
	{Format: "02/01/2006", Description: "UK format", Example: "02/01/2006"},
	{Format: "Jan 2, 06", Description: "Short month with year", Example: "Jan 2, 06"},
	{Format: "Jan 2, 2006", Description: "Long month with year", Example: "Jan 2, 2006"},
	{Format: "01/02/06", Description: "Short US format", Example: "01/02/06"},
	{Format: "02/01/06", Description: "Short UK format", Example: "02/01/06"},
	{Format: "06/02/01", Description: "Short reversed format", Example: "06/02/01"},
	{Format: "06/01/02", Description: "Short alternate format", Example: "06/01/02"},
	{Format: "06-Jan-02", Description: "Short month with dash", Example: "06-Jan-02"},
	{Format: "01/06", Description: "Month/Year only", Example: "01/06"},
	{Format: "02/06", Description: "Alternate Month/Year", Example: "02/06"},
	{Format: "Jan-06", Description: "Short month-year", Example: "Jan-06"},
}

// ParseDateWithFormats mencoba parse tanggal dengan multiple format
func ParseDateWithFormats(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// Menangani format khusus "Feb-24" -> "Feb-2024"
	if strings.Contains(dateStr, "-") && len(dateStr) == 5 {
		dateStr = dateStr[:3] + "20" + dateStr[4:]
	}

	for _, format := range CommonDateFormats {
		parsedDate, err := time.Parse(format.Format, dateStr)
		if err == nil {
			return &parsedDate, nil
		}
	}

	return nil, fmt.Errorf("tidak dapat memparse tanggal: %s", dateStr)
}

// FormatDate memformat tanggal ke format yang diinginkan
func FormatDate(date time.Time, format string) string {
	return date.Format(format)
}

// ParseDateOrDefault mencoba parse tanggal, return default value jika gagal
func ParseDateOrDefault(dateStr string, defaultValue time.Time) time.Time {
	parsed, err := ParseDateWithFormats(dateStr)
	if err != nil || parsed == nil {
		return defaultValue
	}
	return *parsed
}

// IsValidDate mengecek apakah string bisa diparsing sebagai tanggal
func IsValidDate(dateStr string) bool {
	_, err := ParseDateWithFormats(dateStr)
	return err == nil
}

// GetMonthYear mengembalikan bulan dan tahun dari tanggal
func GetMonthYear(date time.Time) string {
	return date.Format("January 2006")
}

// AddCustomFormat menambahkan format tanggal kustom
func AddCustomFormat(format DateFormat) {
	CommonDateFormats = append(CommonDateFormats, format)
}

type ExcelImportConfig struct {
	SheetName   string
	MinColumns  int
	HeaderRows  int // Untuk skip baris header
	ProcessRow  func(row []string, rowIndex int) error
	LogProgress bool // Untuk mengontrol logging
}

func ImportExcelFile(c *gin.Context, config ExcelImportConfig) error {
	if config.LogProgress {
		log.Println("Starting Excel Import function")
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		return fmt.Errorf("error retrieving file: %v", err)
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "*.xlsx")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := io.Copy(tempFile, file); err != nil {
		log.Printf("Error copying file: %v", err)
		return fmt.Errorf("error copying file: %v", err)
	}

	tempFile.Seek(0, 0)
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(config.SheetName)
	if err != nil {
		log.Printf("Error getting rows: %v", err)
		return fmt.Errorf("error getting rows: %v", err)
	}

	if config.LogProgress {
		log.Printf("Total rows found: %d", len(rows))
	}

	for i, row := range rows {
		if i < config.HeaderRows {
			if config.LogProgress {
				log.Printf("Skipping row %d (header rows)", i+1)
			}
			continue
		}

		nonEmptyCount := 0
		for _, cell := range row {
			if cell != "" {
				nonEmptyCount++
			}
		}

		if nonEmptyCount < config.MinColumns {
			if config.LogProgress {
				log.Printf("Row %d skipped: less than %d columns filled, only %d filled",
					i+1, config.MinColumns, nonEmptyCount)
			}
			continue
		}

		if config.LogProgress {
			log.Printf("Processing row %d", i+1)
		}

		if err := config.ProcessRow(row, i); err != nil {
			log.Printf("Error processing row %d: %v", i+1, err)
			continue
		}

		if config.LogProgress {
			log.Printf("Row %d processed successfully", i+1)
		}
	}

	if config.LogProgress {
		log.Println("Excel Import function completed")
	}

	return nil
}

// ********** END OF IMPORT EXCEL ********** //
