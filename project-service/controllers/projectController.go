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
	"github.com/arkaramadhan/its-vo/project-service/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ProjectRequest struct {
	ID              uint    `gorm:"primaryKey"`
	KodeProject     *string `json:"kode_project"`
	JenisPengadaan  *string `json:"jenis_pengadaan"`
	NamaPengadaan   *string `json:"nama_pengadaan"`
	DivInisiasi     *string `json:"div_inisiasi"`
	Bulan           *string `json:"bulan"`
	SumberPendanaan *string `json:"sumber_pendanaan"`
	Anggaran        *string `json:"anggaran"`
	NoIzin          *string `json:"no_izin"`
	TanggalIzin     *string `json:"tanggal_izin"`
	TanggalTor      *string `json:"tanggal_tor"`
	Pic             *string `json:"pic"`
	CreateBy        string  `json:"create_by"`
	Group           *string `json:"group"`
	InfraType       *string `json:"infra_type"`
	BudgetType      *string `json:"budget_type"`
	Type            *string `json:"type"`
}

func UploadHandlerProject(c *gin.Context) {
	helper.UploadHandler(c, "/app/UploadedFile/project")
}

func GetFilesByIDProject(c *gin.Context) {
	helper.GetFilesByID(c)
}

func DeleteFileHandlerProject(c *gin.Context) {
	helper.DeleteFileHandler(c, "/app/UploadedFile/project")
}

func DownloadFileHandlerProject(c *gin.Context) {
	helper.DownloadFileHandler(c, "/app/UploadedFile/project")
}

func ProjectCreate(c *gin.Context) {
	var requestBody ProjectRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Tambahkan pemeriksaan nil untuk Group
	if requestBody.Group == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Group is required"})
		return
	}

	// Generate KodeProject based on Group
	var lastNumber int
	var newKodeProject string
	currentYear := time.Now().Format("2006")

	// Fetch the last project of the same group and year
	lastProject := models.Project{}
	initializers.DB.Where("kode_project LIKE ?", fmt.Sprintf("%%/%s/%%/%s", *requestBody.Group, currentYear)).Order("id desc").First(&lastProject)

	if lastProject.KodeProject != nil {
		fmt.Sscanf(*lastProject.KodeProject, "%d/", &lastNumber)
	}

	newNumber := fmt.Sprintf("%05d", lastNumber+1) // Increment and format
	newKodeProject = fmt.Sprintf("%s/%s/%s/%s/%s/%s", newNumber, *requestBody.Group, *requestBody.InfraType, *requestBody.BudgetType, *requestBody.Type, currentYear)
	requestBody.KodeProject = &newKodeProject

	var bulan *time.Time
	if requestBody.Bulan != nil && *requestBody.Bulan != "" {
		parsedBulan, err := time.Parse("2006-01", *requestBody.Bulan)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			c.JSON(400, gin.H{"message": "Invalid format bulan: " + err.Error()})
			return
		}
		bulan = &parsedBulan
	}

	log.Printf("Parsed date: %v", bulan)

	var tanggal_izin *time.Time
	if requestBody.TanggalIzin != nil && *requestBody.TanggalIzin != "" {
		parsedTanggalIzin, err := time.Parse("2006-01-02", *requestBody.TanggalIzin)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			c.JSON(400, gin.H{"message": "Invalid format tanggal izin: " + err.Error()})
			return
		}
		tanggal_izin = &parsedTanggalIzin
	}

	log.Printf("Parsed date: %v", tanggal_izin)

	var tanggal_tor *time.Time
	if requestBody.TanggalTor != nil && *requestBody.TanggalTor != "" {
		parsedTanggalTor, err := time.Parse("2006-01-02", *requestBody.TanggalTor)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			c.JSON(400, gin.H{"message": "Invalid format tanggal tor: " + err.Error()})
			return
		}
		tanggal_tor = &parsedTanggalTor
	}

	log.Printf("Parsed date: %v", tanggal_tor)

	requestBody.CreateBy = c.MustGet("username").(string)

	project := models.Project{
		KodeProject:     requestBody.KodeProject,
		JenisPengadaan:  requestBody.JenisPengadaan,
		NamaPengadaan:   requestBody.NamaPengadaan,
		DivInisiasi:     requestBody.DivInisiasi,
		Bulan:           bulan,
		SumberPendanaan: requestBody.SumberPendanaan,
		Anggaran:        requestBody.Anggaran,
		NoIzin:          requestBody.NoIzin,
		TanggalIzin:     tanggal_izin,
		TanggalTor:      tanggal_tor,
		Pic:             requestBody.Pic,
		Group:           requestBody.Group,
		InfraType:       requestBody.InfraType,
		BudgetType:      requestBody.BudgetType,
		Type:            requestBody.Type,
		CreateBy:        requestBody.CreateBy,
	}

	// Log data project yang baru dibuat
	log.Printf("Creating new project: %+v", project)

	result := initializers.DB.Create(&project)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal membuat project: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project berhasil dibuat", "project": project})
}

func ProjectIndex(c *gin.Context) {
	var projects []models.Project
	helper.FetchAllRecords(initializers.DB, c, &projects, "project.projects", "Gagal mengambil data project")
}

func ProjectShow(c *gin.Context) {
	id := c.Params.ByName("id")
	var bc models.Project
	helper.ShowRecord(c, initializers.DB, id, &bc, "project berhasil dilihat", "project.projects")
}

func ProjectUpdate(c *gin.Context) {
	var requestBody ProjectRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	id := c.Params.ByName("id")
	var project models.Project
	if err := initializers.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Project tidak ditemukan"})
		return
	}

	// Update KodeProject if Group or other relevant fields are changed
	currentYear := time.Now().Format("2006")
	var group, infraType, budgetType, projectType string

	if requestBody.Group != nil {
		group = *requestBody.Group
	} else {
		group = *project.Group
	}

	if requestBody.InfraType != nil {
		infraType = *requestBody.InfraType
	} else {
		infraType = *project.InfraType
	}

	if requestBody.BudgetType != nil {
		budgetType = *requestBody.BudgetType
	} else {
		budgetType = *project.BudgetType
	}

	if requestBody.Type != nil {
		projectType = *requestBody.Type
	} else {
		projectType = *project.Type
	}

	lastProject := models.Project{}
	initializers.DB.Where("kode_project LIKE ?", fmt.Sprintf("%%/%s/%%/%s", group, currentYear)).Order("id desc").First(&lastProject)
	var lastNumber int
	if lastProject.KodeProject != nil {
		fmt.Sscanf(*lastProject.KodeProject, "%d/", &lastNumber)
	}
	newNumber := fmt.Sprintf("%05d", lastNumber+1)
	newKodeProject := fmt.Sprintf("%s/%s/%s/%s/%s/%s", newNumber, group, infraType, budgetType, projectType, currentYear)
	project.KodeProject = &newKodeProject

	if requestBody.Bulan != nil && *requestBody.Bulan != "" {
		parsedBulan, err := time.Parse("2006-01-02", *requestBody.Bulan)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid format bulan: " + err.Error()})
			return
		}
		project.Bulan = &parsedBulan
	}

	if requestBody.TanggalIzin != nil && *requestBody.TanggalIzin != "" {
		parsedTanggal_izin, err := time.Parse("2006-01-02", *requestBody.TanggalIzin)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid format tanggal izin: " + err.Error()})
			return
		}
		project.TanggalIzin = &parsedTanggal_izin
	}

	if requestBody.TanggalTor != nil && *requestBody.TanggalTor != "" {
		parsedTanggal_tor, err := time.Parse("2006-01-02", *requestBody.TanggalTor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid format tanggal tor: " + err.Error()})
			return
		}
		project.TanggalTor = &parsedTanggal_tor
	}

	// Update other fields
	if requestBody.JenisPengadaan != nil {
		project.JenisPengadaan = requestBody.JenisPengadaan
	}
	if requestBody.NamaPengadaan != nil {
		project.NamaPengadaan = requestBody.NamaPengadaan
	}
	if requestBody.DivInisiasi != nil {
		project.DivInisiasi = requestBody.DivInisiasi
	}
	if requestBody.SumberPendanaan != nil {
		project.SumberPendanaan = requestBody.SumberPendanaan
	}
	if requestBody.Anggaran != nil {
		project.Anggaran = requestBody.Anggaran
	}
	if requestBody.NoIzin != nil {
		project.NoIzin = requestBody.NoIzin
	}
	if requestBody.Pic != nil {
		project.Pic = requestBody.Pic
	}
	project.CreateBy = c.MustGet("username").(string)

	// Save changes
	if err := initializers.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal mengupdate project: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project berhasil diupdate"})
}

func ProjectDelete(c *gin.Context) {
	var project models.Project
	helper.DeleteRecordByID(c, initializers.DB, "project.projects", &project, "project")
}

func ImportExcelProject(c *gin.Context) {
	log.Println("Starting ImportExcelProject function")

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving the file: %v", err)
		c.String(http.StatusBadRequest, "Error retrieving the file: %v", err)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "*.xlsx")
	if err != nil {
		log.Printf("Error creating temporary file: %v", err)
		c.String(http.StatusInternalServerError, "Error creating temporary file: %v", err)
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := io.Copy(tempFile, file); err != nil {
		log.Printf("Error copying file: %v", err)
		c.String(http.StatusInternalServerError, "Error copying file: %v", err)
		return
	}

	tempFile.Seek(0, 0)
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		log.Printf("Error opening file: %v", err)
		c.String(http.StatusInternalServerError, "Error opening file: %v", err)
		return
	}
	defer f.Close()

	sheetName := "PROJECT"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Printf("Error getting rows: %v", err)
		c.String(http.StatusInternalServerError, "Error getting rows: %v", err)
		return
	}

	log.Printf("Total rows found: %d", len(rows))

	for i, row := range rows {
		if i < 6 { // Skip header or initial rows if necessary
			log.Printf("Skipping row %d (header or initial rows)", i+1)
			continue
		}
		// Count non-empty columns
		nonEmptyCount := 0
		for _, cell := range row {
			if cell != "" {
				nonEmptyCount++
			}
		}

		// Skip rows with less than 3 non-empty columns
		if nonEmptyCount < 3 {
			log.Printf("Row %d skipped: less than 3 columns filled, only %d filled", i+1, nonEmptyCount)
			continue
		}

		// Membersihkan string anggaran dari karakter non-numerik
		rawAnggaran := (getColumn(row, 7))
		var anggaranCleaned *string
		if rawAnggaran != "" {
			cleanedAnggaran := helper.CleanNumericString(rawAnggaran)
			anggaranCleaned = &cleanedAnggaran
		}

		project := models.Project{
			KodeProject:     helper.GetStringOrNil(getColumn(row, 1)),
			JenisPengadaan:  helper.GetStringOrNil(getColumn(row, 2)),
			NamaPengadaan:   helper.GetStringOrNil(getColumn(row, 3)),
			DivInisiasi:     helper.GetStringOrNil(getColumn(row, 4)),
			Bulan:           helper.ParseDateOrNil(helper.GetStringOrNil(getColumn(row, 5))),
			SumberPendanaan: helper.GetStringOrNil(getColumn(row, 6)),
			Anggaran:        anggaranCleaned,
			NoIzin:          helper.GetStringOrNil(getColumn(row, 8)),
			TanggalIzin:     helper.ParseDateOrNil(helper.GetStringOrNil(getColumn(row, 9))),
			TanggalTor:      helper.ParseDateOrNil(helper.GetStringOrNil(getColumn(row, 10))),
			Pic:             helper.GetStringOrNil(getColumn(row, 11)),
			CreateBy:        c.MustGet("username").(string),
		}

		// Log data yang diimpor
		log.Printf("Importing row %d", i+1)

		if err := initializers.DB.Create(&project).Error; err != nil {
			log.Printf("Error saving record from row %d: %v", i+1, err)
			continue
		} else {
			log.Printf("Record from row %d saved successfully", i+1)
		}
	}

	log.Println("ImportExcelProject function completed")
	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diimport"})
}

// Helper function to safely get column data or return empty if index is out of range
func getColumn(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return row[index]
}

// Helper function to parse date from various formats
func parseDate(dateStr string) (time.Time, error) {
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

func ExportProjectHandler(c *gin.Context) {
	f := excelize.NewFile()

	ExportProjectToExcel(c, f, "PROJECT", true)
}

func ExportProjectToExcel(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error {
	// 1. Ambil data dari database
	var projects []models.Project
	initializers.DB.Table("project.projects").Find(&projects)

	// 2. Konversi ke interface ExcelData
	var excelData []helper.ExcelData
	for _, p := range projects {
		excelData = append(excelData, &p)
	}

	// 3. Siapkan konfigurasi
	config := helper.ExcelConfig{
		SheetName: "PROJECT",
		Columns: []helper.ExcelColumn{
			{Header: "Kode Project", Width: 38},
			{Header: "Jenis Pengadaan", Width: 27},
			{Header: "Nama Pengadaan", Width: 40},
			{Header: "Divisi Inisiasi", Width: 20},
			{Header: "Bulan", Width: 10},
			{Header: "Sumber Pendanaan", Width: 20},
			{Header: "Anggaran", Width: 20},
			{Header: "No Izin", Width: 23},
			{Header: "Tgl Izin", Width: 14},
			{Header: "Tgl TOR", Width: 14},
			{Header: "Pic", Width: 16},
		},
		Data:         excelData,
		IsSplitSheet: true,
		GetStatus:    nil,
		SplitType:    helper.SplitHorizontal,
		CustomStyles: &helper.CustomStyles{
			SeparatorLabels: map[string]string{
				"ITS-SAG": "SAG",
				"ITS-ISO": "ISO",
			},
			SeparatorStyle: &excelize.Style{
				Fill:      excelize.Fill{Type: "pattern", Color: []string{"000000"}, Pattern: 1},
				Font:      &excelize.Font{Bold: true, Color: "FFFFFF", VertAlign: "center"},
				Alignment: helper.CenterAlignment,
				Border:    helper.BorderBlack,
			},
			DefaultCellStyle: &excelize.Style{
				Alignment: helper.WrapAlignment,
				Border:    helper.BorderBlack,
			},
			AnggaranStyle: &excelize.Style{
				NumFmt: 3,
				Border: helper.BorderBlack,
			},
		},
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
