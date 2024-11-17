package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/models"
	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "File diperlukan"})
		return
	}

	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	dir := filepath.Join(baseDir, id)
	// if _, err := os.Stat(dir); os.IsNotExist(err) {
	// 	os.MkdirAll(dir, 0755)
	// }

	filePath := filepath.Join(dir, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan metadata file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File berhasil diunggah"})
}

func GetFilesByID(c *gin.Context) {
	id := c.Param("id")

	var files []models.File
	result := initializers.DB.Table("common.files").Where("user_id = ?", id).Find(&files)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data file"})
		return
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.FileName)
	}

	c.JSON(http.StatusOK, gin.H{"files": fileNames})
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
	log.Printf("Received ID: %s and Filename: %s", id, filename) // Tambahkan log ini

	baseDir := mainDir
	fullPath := filepath.Join(baseDir, id, filename)

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

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
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
	var searchPattern string
	if docType == "perdin" {
		searchPattern = fmt.Sprintf("%%/%s/%d", category, currentYear)
	} else {
		searchPattern = fmt.Sprintf("%%/ITS-%s/%s/%d", category, docType, currentYear) // Ini akan mencari format seperti '%/ITS-SAG/M/2023", category, docType)
	}

	// Log untuk debugging
	log.Printf("Mencari dokumen dengan pattern: %s di schema: %s", searchPattern, schema)

	result := initializers.DB.Table(schema).
		Where(fmt.Sprintf("%s LIKE ?", dbField), searchPattern).
		Order("id desc").
		First(model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Ini bukan error, ini expected behavior untuk dokumen pertama
			var newNumber string
			if docType == "perdin" {
				newNumber = fmt.Sprintf("00001/%s/%d", category, currentYear)
			} else {
				newNumber = fmt.Sprintf("00001/ITS-%s/%s/%d", category, docType, currentYear)
			}
			log.Printf("Tidak ada dokumen sebelumnya, membuat nomor pertama: %s", newNumber)
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

	newNumber := fmt.Sprintf("%05d/ITS-%s/%s/%d", num+1, category, docType, currentYear)
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

func CreateRecord[T any](db *gorm.DB, c *gin.Context, input *T, createByField *string) {
	// Langsung bind ke struct
	if err := c.ShouldBindJSON(input); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Set CreateBy
	if createByField != nil {
		if creator, exists := c.Get("username"); exists {
			*createByField = creator.(string)
		}
	}

	// Simpan ke database
	if err := db.Table("dokumen.memos").Create(input).Error; err != nil {
		log.Printf("Error saving record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create record"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func ShowRecord[T any](c *gin.Context, db *gorm.DB, id string, data *T, successMessage string, schema string) {
	if err := db.Table(schema).First(data, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "record tidak ditemukan"})
		return
	}
	log.Printf("Data retrieved: %+v", data) // Tambahkan log ini untuk melihat data yang diambil
	c.JSON(http.StatusOK, data)
}

type DocumentUpdateParams struct {
	DB              *gorm.DB
	C               *gin.Context
	Request         interface{}
	Document        interface{}
	DocumentType    string
	GetLatestNumber func(string) (string, error)
}

type DynamicDocument interface {
	GetProperty(key string) interface{}
	SetProperty(key string, value interface{}) error
}

func UpdateDocument(params DocumentUpdateParams) error {
	var err error
	document, ok := params.Document.(DynamicDocument)
	if !ok {
		RespondError(params.C, http.StatusBadRequest, "Document does not implement DynamicDocument interface")
		return fmt.Errorf("Document does not implement DynamicDocument interface")
	}

	// Bind JSON request body
	if err = params.C.BindJSON(&params.Request); err != nil {
		log.Printf("Error binding JSON for %s: %v", params.DocumentType, err)
		RespondError(params.C, http.StatusBadRequest, "invalid request body: "+err.Error())
		return fmt.Errorf("invalid request body: %v", err)
	}

	// Ambil ID dari parameter URL
	id := params.C.Param("id")

	// Cari dokumen berdasarkan ID
	if err = params.DB.First(document, id).Error; err != nil {
		log.Printf("Error finding %s with ID %s: %v", params.DocumentType, id, err)
		RespondError(params.C, http.StatusNotFound, fmt.Sprintf("%s not found", params.DocumentType))
		return fmt.Errorf("Error finding %s with ID %s: %v", params.DocumentType, id, err)
	}
	// Proses pembaruan spesifik dokumen
	requestMap, ok := params.Request.(map[string]interface{})
	if !ok {
		RespondError(params.C, http.StatusBadRequest, "Request body must be a JSON object")
		return fmt.Errorf("Request body must be a JSON object")
	}
	if err = processDocumentUpdate(document, requestMap); err != nil {
		log.Printf("Error updating %s: %v", params.DocumentType, err)
		RespondError(params.C, http.StatusInternalServerError, err.Error())
		return fmt.Errorf("Error updating %s: %v", params.DocumentType, err)
	}

	// Simpan perubahan
	if err = params.DB.Save(document).Error; err != nil {
		log.Printf("Error saving updated %s: %v", params.DocumentType, err)
		RespondError(params.C, http.StatusInternalServerError, fmt.Sprintf("failed to update %s: %s", params.DocumentType, err.Error()))
		return fmt.Errorf("Error saving updated %s: %v", params.DocumentType, err)
	}

	// Tambahkan log untuk mencatat data yang diupdate
	log.Printf("Data updated successfully: %v", document)

	// Return response sukses
	params.C.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s successfully updated", params.DocumentType)})
	return nil
}

func processDocumentUpdate(document DynamicDocument, request map[string]interface{}) error {
	// Implementasi logika update menggunakan interface
	for key, value := range request {
		if err := document.SetProperty(key, value); err != nil {
			return err
		}
	}
	return nil
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

	if err := db.Table(schema).First(model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": modelName + " tidak ditemukan"})
		return
	}

	if err := db.Table(schema).Delete(model).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "gagal menghapus " + modelName + ": " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": modelName + " berhasil dihapus"})
}
