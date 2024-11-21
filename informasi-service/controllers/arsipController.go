package controllers

import (
	"net/http"
	"time"

	"github.com/arkaramadhan/its-vo/common/initializers"
	helper "github.com/arkaramadhan/its-vo/common/utils"
	"github.com/arkaramadhan/its-vo/informasi-service/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type arsipRequest struct {
	ID                uint    `gorm:"primaryKey"`
	NoArsip           *string `json:"no_arsip"`
	JenisDokumen      *string `json:"jenis_dokumen"`
	NoDokumen         *string `json:"no_dokumen"`
	Perihal           *string `json:"perihal"`
	NoBox             *string `json:"no_box"`
	TanggalDokumen    *string `json:"tanggal_dokumen"`
	TanggalPenyerahan *string `json:"tanggal_penyerahan"`
	Keterangan        *string `json:"keterangan"`
	CreateBy          string  `json:"create_by"`
}

func UploadHandlerArsip(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/arsip")
}

func GetFilesByIDArsip(c *gin.Context) {
	helper.GetFilesByID(c, "/app/UploadedFile/arsip")
}

func DeleteFileHandlerArsip(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/arsip")
}

func DownloadFileHandlerArsip(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/arsip")
}

func ArsipIndex(c *gin.Context) {
	var arsips []models.Arsip
	helper.FetchAllRecords(initializers.DB, c, &arsips, "informasi.arsips", "Gagal mengambil data arsip")
}

// Fungsi untuk membuat arsip baru
func ArsipCreate(c *gin.Context) {
	var requestBody arsipRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
		return
	}
	requestBody.CreateBy = c.MustGet("username").(string)

	var tanggal *time.Time
	if requestBody.TanggalDokumen != nil && *requestBody.TanggalDokumen != "" {
		parsedTanggal, err := time.Parse("2006-01-02", *requestBody.TanggalDokumen)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid format tanggal: " + err.Error()})
			return
		}
		tanggal = &parsedTanggal
	}

	arsip := models.Arsip{
		NoArsip:           requestBody.NoArsip,
		JenisDokumen:      requestBody.JenisDokumen,
		NoDokumen:         requestBody.NoDokumen,
		Perihal:           requestBody.Perihal,
		NoBox:             requestBody.NoBox,
		Keterangan:        requestBody.Keterangan,
		TanggalDokumen:    tanggal,
		TanggalPenyerahan: tanggal, // Assuming same date handling for TanggalPenyerahan
		CreateBy:          requestBody.CreateBy,
	}

	if err := initializers.DB.Create(&arsip).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal membuat arsip: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, arsip)
}

func ArsipShow(c *gin.Context) {
	id := c.Param("id")
	var arsip models.Arsip
	if err := initializers.DB.First(&arsip, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Arsip tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"arsip": arsip})
}

func ArsipUpdate(c *gin.Context) {
	id := c.Param("id")
	var requestBody arsipRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
		return
	}

	var arsip models.Arsip
	if err := initializers.DB.First(&arsip, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Arsip tidak ditemukan"})
		return
	}

	if requestBody.TanggalDokumen != nil {
		tanggal, err := time.Parse("2006-01-02", *requestBody.TanggalDokumen)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid format tanggal: " + err.Error()})
			return
		}
		arsip.TanggalDokumen = &tanggal
	}

	// Update fields if provided in request
	if requestBody.NoArsip != nil {
		arsip.NoArsip = requestBody.NoArsip
	}
	if requestBody.JenisDokumen != nil {
		arsip.JenisDokumen = requestBody.JenisDokumen
	}
	if requestBody.NoDokumen != nil {
		arsip.NoDokumen = requestBody.NoDokumen
	}
	if requestBody.Perihal != nil {
		arsip.Perihal = requestBody.Perihal
	}
	if requestBody.NoBox != nil {
		arsip.NoBox = requestBody.NoBox
	}
	if requestBody.Keterangan != nil {
		arsip.Keterangan = requestBody.Keterangan
	}
	if requestBody.CreateBy != "" {
		arsip.CreateBy = requestBody.CreateBy
	}

	if err := initializers.DB.Save(&arsip).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal mengupdate arsip: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "arsip berhasil diupdate"})
}

func ArsipDelete(c *gin.Context) {
	id := c.Param("id")
	var arsip models.Arsip
	if err := initializers.DB.Where("id = ?", id).Delete(&arsip).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal menghapus arsip: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Arsip berhasil dihapus"})
}

func ExportArsipHandler(c *gin.Context) {
	f := excelize.NewFile()
	ExportArsipToExcel(c, f, "ARSIP", true)
}

func ExportArsipToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var arsip []models.Arsip
	initializers.DB.Table("informasi.arsips").Find(&arsip)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, asp := range arsip {
		excelData = append(excelData, &asp)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "ARSIP",
		Columns: []helper.ExcelColumn{
			{Header: "No Arsip", Width: 20},
			{Header: "Jenis Dokumen", Width: 27},
			{Header: "No Dokumen", Width: 40},
			{Header: "Perihal", Width: 20},
			{Header: "No Box", Width: 20},
			{Header: "Tanggal Dokumen", Width: 20},
			{Header: "Tanggal Penyerahan", Width: 20},
			{Header: "Keterangan", Width: 20},
		},
		Data:         excelData,
		IsSplitSheet: false,
		GetStatus:    nil,
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

func ImportExcelArsip(c *gin.Context) {
	config := helper.ExcelImportConfig{
		SheetName:   "ARSIP",
		MinColumns:  2,
		HeaderRows:  1,
		LogProgress: true,
		ProcessRow: func(row []string, rowIndex int) error {
			// Ambil data dari kolom
			noArsip := helper.GetColumn(row, 0)
			jenisDokumen := helper.GetColumn(row, 1)
			noDokumen := helper.GetColumn(row, 2)
			perihal := helper.GetColumn(row, 3)
			noBox := helper.GetColumn(row, 4)
			tanggalDocStr := helper.GetColumn(row, 5)
			tanggalPenyerahanStr := helper.GetColumn(row, 6)
			keterangan := helper.GetColumn(row, 7)

			// Parse tanggal
			tanggalDoc, _ := helper.ParseDateWithFormats(tanggalDocStr)
			tanggalPenyerahan, _ := helper.ParseDateWithFormats(tanggalPenyerahanStr)

			// Buat dan simpan memo
			arsip := models.Arsip{
				TanggalDokumen:    tanggalDoc,
				TanggalPenyerahan: tanggalPenyerahan,
				NoArsip:           &noArsip,
				JenisDokumen:      &jenisDokumen,
				NoDokumen:         &noDokumen,
				Perihal:           &perihal,
				NoBox:             &noBox,
				Keterangan:        &keterangan,
				CreateBy:          c.MustGet("username").(string),
			}

			return initializers.DB.Create(&arsip).Error
		},
	}

	if err := helper.ImportExcelFile(c, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}
