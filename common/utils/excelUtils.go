package utils

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

var (
	BorderBlack = []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	}

	centerAlignment = &excelize.Alignment{
		Horizontal: "center",
		Vertical:   "center",
	}

	wrapAlignment = &excelize.Alignment{
		WrapText: true,
	}

	FontBlack = &excelize.Font{
		Color: "000000",
		Bold:  true,
	}
)

type SplitType int

const (
	SplitHorizontal SplitType = iota
	SplitVertical
)

// ExcelData adalah interface yang harus diimplementasikan oleh semua model
type ExcelData interface {
	ToExcelRow() []interface{}
	GetDocType() string // Untuk menentukan tipe dokumen (SAG/ISO)
}

type CustomStyles struct {
	StatusStyles     map[string]*excelize.Style
	DefaultCellStyle *excelize.Style
	SeparatorStyle   *excelize.Style
	SeparatorLabels  map[string]string
	AnggaranStyle    *excelize.Style
}

// ExcelColumn mendefinisikan struktur untuk setiap kolom Excel
type ExcelColumn struct {
	Header string
	Width  float64
}

type GetStatusFunc func(data interface{}) string

// ExcelConfig menyimpan konfigurasi untuk ekspor Excel
type ExcelConfig struct {
	SheetName    string
	Columns      []ExcelColumn
	Data         []ExcelData
	FileName     string
	IsSplitSheet bool // true jika perlu dipisah SAG/ISO
	CustomStyles *CustomStyles
	GetStatus    GetStatusFunc
	SplitType    SplitType
}

func ExportToExcel(config ExcelConfig) (*excelize.File, error) {
	f := excelize.NewFile()

	f.NewSheet(config.SheetName)
	f.DeleteSheet("Sheet1")

	// Set headers dan style
	styleHeader, err := CreateHeaderStyle(f)
	if err != nil {
		return nil, err
	}

	// Tambahkan pengecekan untuk SeparatorStyle
	if config.CustomStyles != nil && config.CustomStyles.SeparatorStyle != nil {
		styleID, err := f.NewStyle(config.CustomStyles.SeparatorStyle)
		if err != nil {
			return nil, err
		}
		// Terapkan style untuk baris pemisah SAG
		f.SetCellValue(config.SheetName, "F2", "SAG")
		f.SetCellStyle(config.SheetName, "A2", fmt.Sprintf("%s2", string(rune('A'+len(config.Columns)-1))), styleID)
	}

	if config.CustomStyles != nil && config.CustomStyles.DefaultCellStyle != nil {
		styleID, _ := f.NewStyle(config.CustomStyles.DefaultCellStyle)
		startCell := fmt.Sprintf("A%d", 3) // Mulai dari baris 3 setelah header dan pemisah
		endCell := fmt.Sprintf("%s%d", string(rune('A'+len(config.Columns)-1)), 3)
		f.SetCellStyle(config.SheetName, startCell, endCell, styleID)
	}

	if config.IsSplitSheet {
		return ExportSplitSheet(f, config, styleHeader)
	}
	return ExportSingleSheet(f, config, styleHeader)
}

func ExportToSheet(f *excelize.File, config ExcelConfig) error {
	// Buat sheet baru
	f.NewSheet(config.SheetName)
	f.DeleteSheet("Sheet1")

	// Set headers dan style
	styleHeader, err := CreateHeaderStyle(f)
	if err != nil {
		return err
	}

	if config.CustomStyles != nil && config.CustomStyles.StatusStyles != nil {
		status := config.GetStatus(config.Data[0])
		if style, ok := config.CustomStyles.StatusStyles[status]; ok {
			statusCell := fmt.Sprintf("%s%d", string(rune('A'+len(config.Columns)-1)), 2)
			styleID, err := f.NewStyle(style)
			if err != nil {
				return err
			}
			f.SetCellStyle(config.SheetName, statusCell, statusCell, styleID)
		}
	}

	if config.CustomStyles != nil && config.CustomStyles.DefaultCellStyle != nil {
		styleID, _ := f.NewStyle(config.CustomStyles.DefaultCellStyle)
		// Apply to all cells in the row except status column
		startCell := fmt.Sprintf("A%d", 2)
		endCell := fmt.Sprintf("%s%d", string(rune('A'+len(config.Columns)-2)), 2)
		f.SetCellStyle(config.SheetName, startCell, endCell, styleID)
	}

	if config.IsSplitSheet {
		_, err = ExportSplitSheet(f, config, styleHeader)
	} else {
		_, err = ExportSingleSheet(f, config, styleHeader)
	}

	if err != nil {
		return err
	}

	return nil
}

func ExportSingleSheet(f *excelize.File, config ExcelConfig, styleHeader int) (*excelize.File, error) {
	// Set headers dan lebar kolom
	for i, col := range config.Columns {
		cell := fmt.Sprintf("%s1", string('A'+i))
		f.SetCellValue(config.SheetName, cell, col.Header)
		f.SetColWidth(config.SheetName, string('A'+i), string('A'+i), col.Width)
		f.SetCellStyle(config.SheetName, cell, cell, styleHeader)
	}

	// Proses data
	row := 2
	for _, item := range config.Data {
		values := item.ToExcelRow()
		for col, value := range values {
			cell := fmt.Sprintf("%s%d", string('A'+col), row)
			f.SetCellValue(config.SheetName, cell, value)

			// Terapkan style berdasarkan kolom status
			if col == 2 && config.CustomStyles != nil { // Asumsikan kolom status adalah kolom ke-3 (index 2)
				if config.GetStatus != nil {
					status := config.GetStatus(item)
					if style, ok := config.CustomStyles.StatusStyles[status]; ok {
						styleID, err := f.NewStyle(style)
						if err != nil {
							return nil, err
						}
						f.SetCellStyle(config.SheetName, cell, cell, styleID)
					}
				}
			} else if config.CustomStyles != nil && config.CustomStyles.DefaultCellStyle != nil {
				// Terapkan default style untuk kolom non-status
				styleID, err := f.NewStyle(config.CustomStyles.DefaultCellStyle)
				if err != nil {
					return nil, err
				}
				f.SetCellStyle(config.SheetName, cell, cell, styleID)
			}
		}
		row++
	}

	return f, nil
}

func ExportSplitSheet(f *excelize.File, config ExcelConfig, styleHeader int) (*excelize.File, error) {
	if config.SplitType == SplitVertical {
		return exportVerticalSplit(f, config, styleHeader)
	}
	return exportHorizontalSplit(f, config, styleHeader)
}

func exportVerticalSplit(f *excelize.File, config ExcelConfig, styleHeader int) (*excelize.File, error) {
	// Set headers SAG (kolom kiri)
	for i, col := range config.Columns {
		cell := fmt.Sprintf("%s1", string('A'+i))
		f.SetCellValue(config.SheetName, cell, col.Header)
		f.SetColWidth(config.SheetName, string('A'+i), string('A'+i), col.Width)
		f.SetCellStyle(config.SheetName, cell, cell, styleHeader)
	}

	// Set headers ISO (kolom kanan)
	for i, col := range config.Columns {
		cell := fmt.Sprintf("%s1", string(byte('F'+i)))
		f.SetCellValue(config.SheetName, cell, col.Header)
		f.SetColWidth(config.SheetName, string(byte('F'+i)), string(byte('F'+i)), col.Width)
		f.SetCellStyle(config.SheetName, cell, cell, styleHeader)
	}

	// Set lebar pemisah
	f.SetColWidth(config.SheetName, "E", "E", 2)

	// Proses data
	rowSAG := 2
	rowISO := 2

	for _, item := range config.Data {
		values := item.ToExcelRow()
		if item.GetDocType() == "SAG" {
			for col, value := range values {
				cell := fmt.Sprintf("%s%d", string('A'+col), rowSAG)
				f.SetCellValue(config.SheetName, cell, value)
			}
			rowSAG++
		} else {
			for col, value := range values {
				cell := fmt.Sprintf("%s%d", string(byte('F'+col)), rowISO)
				f.SetCellValue(config.SheetName, cell, value)
			}
			rowISO++
		}
	}

	// Apply styling
	styleData, err := CreateDataStyle(f)
	if err != nil {
		return nil, err
	}

	lastRow := max(rowSAG, rowISO) - 1
	if lastRow >= 2 {
		// Style untuk kolom SAG
		lastCellSAG := fmt.Sprintf("%s%d", string('A'+len(config.Columns)-1), lastRow)
		f.SetCellStyle(config.SheetName, "A2", lastCellSAG, styleData)

		// Style untuk kolom ISO
		lastCellISO := fmt.Sprintf("%s%d", string(byte('F'+len(config.Columns)-1)), lastRow)
		f.SetCellStyle(config.SheetName, "F2", lastCellISO, styleData)

		// Style untuk garis pemisah
		styleLine, err := CreateLineStyle(f)
		if err != nil {
			return nil, err
		}
		f.SetCellStyle(config.SheetName, "E1", fmt.Sprintf("E%d", lastRow), styleLine)
	}

	return f, nil
}

func exportHorizontalSplit(f *excelize.File, config ExcelConfig, styleHeader int) (*excelize.File, error) {
	// Set headers
	for i, col := range config.Columns {
		cell := fmt.Sprintf("%s1", string('A'+i))
		f.SetCellValue(config.SheetName, cell, col.Header)
		f.SetColWidth(config.SheetName, string('A'+i), string('A'+i), col.Width)
		f.SetCellStyle(config.SheetName, cell, cell, styleHeader)
	}

	// Tambahkan separator SAG di baris 2
	if config.CustomStyles != nil && config.CustomStyles.SeparatorStyle != nil {
		styleID, _ := f.NewStyle(config.CustomStyles.SeparatorStyle)
		f.SetCellValue(config.SheetName, "F2", "SAG")
		f.SetCellStyle(config.SheetName, "A2",
			fmt.Sprintf("%s2", string(rune('A'+len(config.Columns)-1))), styleID)
	}

	rowSAG := 3 // Mulai setelah header dan pemisah SAG
	rowISO := 3
	lastRowSAG := 3

	// Proses data SAG
	for _, item := range config.Data {
		if item.GetDocType() == "SAG" {
			values := item.ToExcelRow()
			for col, value := range values {
				cell := fmt.Sprintf("%s%d", string('A'+col), rowSAG)
				f.SetCellValue(config.SheetName, cell, value)

				// Terapkan style default
				if config.CustomStyles != nil && config.CustomStyles.DefaultCellStyle != nil {
					styleID, _ := f.NewStyle(config.CustomStyles.DefaultCellStyle)
					f.SetCellStyle(config.SheetName, cell, cell, styleID)
				}

				// Style khusus untuk kolom anggaran
				if col == 6 && config.CustomStyles != nil && config.CustomStyles.AnggaranStyle != nil {
					styleID, _ := f.NewStyle(config.CustomStyles.AnggaranStyle)
					f.SetCellStyle(config.SheetName, cell, cell, styleID)
				}
			}
			rowSAG++
			lastRowSAG = rowSAG
		}
	}

	// Tambahkan separator ISO setelah data SAG
	if config.CustomStyles != nil && config.CustomStyles.SeparatorStyle != nil {
		styleID, _ := f.NewStyle(config.CustomStyles.SeparatorStyle)
		f.SetCellValue(config.SheetName, fmt.Sprintf("F%d", lastRowSAG), "ISO")
		f.SetCellStyle(config.SheetName,
			fmt.Sprintf("A%d", lastRowSAG),
			fmt.Sprintf("%s%d", string(rune('A'+len(config.Columns)-1)), lastRowSAG),
			styleID)
		rowISO = lastRowSAG + 1
	}

	// Proses data ISO
	for _, item := range config.Data {
		if item.GetDocType() == "ISO" {
			values := item.ToExcelRow()
			for col, value := range values {
				cell := fmt.Sprintf("%s%d", string('A'+col), rowISO)
				f.SetCellValue(config.SheetName, cell, value)

				// Terapkan style default
				if config.CustomStyles != nil && config.CustomStyles.DefaultCellStyle != nil {
					styleID, _ := f.NewStyle(config.CustomStyles.DefaultCellStyle)
					f.SetCellStyle(config.SheetName, cell, cell, styleID)
				}

				// Style khusus untuk kolom anggaran
				if col == 6 && config.CustomStyles != nil && config.CustomStyles.AnggaranStyle != nil {
					styleID, _ := f.NewStyle(config.CustomStyles.AnggaranStyle)
					f.SetCellStyle(config.SheetName, cell, cell, styleID)
				}
			}
			rowISO++
		}
	}

	// Set tinggi baris
	lastRow := max(rowISO, lastRowSAG)
	for i := 2; i < lastRow; i++ {
		f.SetRowHeight(config.SheetName, i, 30)
	}

	return f, nil

	return f, nil
}

func CreateHeaderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "FFFFFF",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
}

func CreateDataStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
}

func CreateLineStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"000000"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "FFFFFF", Style: 2},
		},
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Helper function untuk mengambil nilai pointer
func GetValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// ExcelEvent adalah interface yang harus di implementasikan oleh model event
type ExcelEvent interface {
	GetTitle() string
	GetStart() time.Time
	GetEnd() time.Time
	GetColor() string
	GetAllDay() bool
	GetResourceID() uint // Opsional, return 0 jika tidak ada
}

type CalenderConfig struct {
	SheetName   string
	FileName    string
	Events      []ExcelEvent
	UseResource bool
	ResourceMap map[uint]string
	RowOffset   int
	ColOffset   int
}

func ExportCalenderToExcel(c *gin.Context, config CalenderConfig) error {
	f := excelize.NewFile()
	sheet := config.SheetName
	f.NewSheet(sheet)

	months := []string{
		"January 2024", "February 2024", "March 2024", "April 2024",
		"May 2024", "June 2024", "July 2024", "August 2024",
		"September 2024", "October 2024", "November 2024", "December 2024",
	}

	for i, month := range months {
		setMonthData(f, sheet, month, config.RowOffset, config.ColOffset, config.Events, config.ResourceMap, config.UseResource)
		config.ColOffset += 9
		if (i+1)%3 == 0 {
			config.RowOffset += 18
			config.ColOffset = 0
		}
	}

	f.DeleteSheet("Sheet1")

	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to write Excel buffer: %v", err)})
		return err
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xlsx", config.FileName))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := c.Writer.Write(buffer.Bytes()); err != nil {
		log.Printf("Error sending Excel file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send Excel file: %v", err)})
		return err
	}
	return nil
}

func ExportCalenderToSheet(f *excelize.File, config CalenderConfig) error {
	// Buat sheet baru
	f.NewSheet(config.SheetName)

	months := []string{
		"January 2024", "February 2024", "March 2024", "April 2024",
		"May 2024", "June 2024", "July 2024", "August 2024",
		"September 2024", "October 2024", "November 2024", "December 2024",
	}
	for i, month := range months {
		err := setMonthData(f, config.SheetName, month, config.RowOffset, config.ColOffset, config.Events, config.ResourceMap, config.UseResource)
		if err != nil {
			return err
		}
		config.ColOffset += 9
		if (i+1)%3 == 0 {
			config.RowOffset += 18
			config.ColOffset = 0
		}
	}

	return nil
}

func setMonthData(f *excelize.File, sheet, month string, rowOffset, colOffset int, events []ExcelEvent, resourceMap map[uint]string, useResource bool) error {
	var (
		monthStyle, titleStyle, dataStyle, blankStyle int
		err                                           error
		addr                                          string
	)

	// Definisikan loc di sini
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		fmt.Printf("Error loading location: %v", err)
		return err
	}

	monthTime, err := time.ParseInLocation("January 2006", month, loc)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Get the first day of the month and the number of days in the month
	firstDay := monthTime.Weekday()
	daysInMonth := time.Date(monthTime.Year(), monthTime.Month()+1, 0, 0, 0, 0, 0, loc).Day()

	// cell values
	data := map[int][]interface{}{
		1 + rowOffset: {month},
		3 + rowOffset: {"MINGGU", "SENIN", "SELASA", "RABU",
			"KAMIS", "JUMAT", "SABTU"},
	}

	// Fill in the dates
	day := 1
	for r := 4 + rowOffset; day <= daysInMonth; r += 2 {
		week := make([]interface{}, 7) // Inisialisasi ulang array week untuk setiap baris baru
		eventDetails := make([]interface{}, 7)
		for d := 0; d < 7; d++ { // Mulai loop dari 0 hingga 6 (Minggu hingga Sabtu)
			if r == 4+rowOffset && d < int(firstDay) {
				// Jika ini adalah baris pertama dan hari ini sebelum 'firstDay', biarkan kosong
				continue
			}
			if day <= daysInMonth {
				week[d] = day // Isi tanggal

				// Cek apakah ada event pada hari ini
				for _, event := range events {
					var startDate, endDate time.Time
					if event.GetAllDay() {

						startDate, err = time.ParseInLocation("2006-01-02", event.GetStart().Format("2006-01-02"), loc)
						if err != nil {
							fmt.Printf("Error parsing start date: %v\n", err)
							continue
						}
						endDate = startDate

					} else {

						startDate, err = time.ParseInLocation("2006-01-02", event.GetStart().Format("2006-01-02"), loc)
						if err != nil {
							fmt.Printf("Error parsing start date: %v\n", err)
							continue
						}
						endDate, err = time.ParseInLocation("2006-01-02", event.GetEnd().Format("2006-01-02"), loc)
						if err != nil {
							fmt.Printf("Error parsing end date: %v\n", err)
							continue
						}

					}
					currentDate := time.Date(monthTime.Year(), monthTime.Month(), day, 0, 0, 0, 0, loc) // Pastikan waktu diatur ke 00:00:00

					// Periksa apakah currentDate sama dengan startDate atau berada di antara startDate dan endDate
					if currentDate.Equal(startDate) || (currentDate.After(startDate) && currentDate.Before(endDate.AddDate(0, 0, 1))) {
						var eventDetail string
						if event.GetAllDay() {
							eventDetail = fmt.Sprintf("%s\nAllDay", event.GetTitle())
						} else if useResource {
							displayText := resourceMap[event.GetResourceID()] // Hanya tampilkan nama resource
							// Cek jika sudah ada teks di sel tersebut, tambahkan dengan koma jika nama resource belum ada
							if existingText, exists := week[d].(string); exists && existingText != "" {
								if !strings.Contains(existingText, displayText) { // Cek apakah nama resource sudah ada
									week[d] = fmt.Sprintf("%s, %s", existingText, displayText)
								}
							} else {
								week[d] = fmt.Sprintf("%d %s", day, displayText) // Tampilkan tanggal hanya pada entri pertama
							}

							eventDetail = event.GetTitle()
						} else {
							startDate = event.GetStart()
							endDate = event.GetEnd()
							FormattedStart := startDate.Format("15:04")
							FormattedEnd := endDate.Format("15:04")
							// Format hanya jam dan menit dari tanggal
							eventDetail = fmt.Sprintf("%s\n%s - %s", event.GetTitle(), FormattedStart, FormattedEnd)
						}

						// Gabungkan detail acara jika sudah ada
						if eventDetails[d] != nil {
							eventDetails[d] = fmt.Sprintf("%s\n%s", eventDetails[d], eventDetail)
						} else {
							eventDetails[d] = eventDetail
						}
						// Hanya terapkan warna untuk event pertama pada hari itu
						if event.GetTitle() != "" && eventDetails[d] == eventDetail { // Perubahan kondisi di sini
							cellAddr, _ := excelize.JoinCellName(string('B'+colOffset+int(d)), r+1)

							// Buat gaya baru dengan warna latar belakang
							style, err := f.NewStyle(&excelize.Style{
								Fill: excelize.Fill{
									Type:    "pattern",
									Color:   []string{event.GetColor()},
									Pattern: 1,
								},
								Font: &excelize.Font{
									Size:  11,
									Color: "FFFFFF",
									Bold:  true,
								},
								Alignment: &excelize.Alignment{WrapText: true},
								Border: []excelize.Border{
									{Type: "left", Color: "DADEE0", Style: 1},
									{Type: "right", Color: "DADEE0", Style: 1},
									{Type: "top", Color: "DADEE0", Style: 1},
									{Type: "bottom", Color: "DADEE0", Style: 1},
								},
							})
							if err != nil {
								fmt.Printf("Error membuat gaya untuk sel %s: %v\n", cellAddr, err)
								continue
							}

							// Terapkan gaya ke sel
							if err := f.SetCellStyle(sheet, cellAddr, cellAddr, style); err != nil {
								fmt.Printf("Error menerapkan gaya ke sel %s: %v\n", cellAddr, err)
							} else {
								fmt.Printf("Berhasil menerapkan gaya dengan warna %s ke sel %s\n", event.GetColor(), cellAddr)
							}
						}
					}
				}

				day++ // Increment day hanya jika hari ini diisi
			}
		}
		data[r] = week
		data[r+1] = eventDetails
		if r == 4+rowOffset {
			firstDay = 0 // Reset firstDay untuk minggu berikutnya
		}
	}

	// custom rows height
	height := map[int]float64{
		1 + rowOffset: 45, 3 + rowOffset: 22, 5 + rowOffset: 30, 7 + rowOffset: 30,
		9 + rowOffset: 30, 11 + rowOffset: 30, 13 + rowOffset: 30, 15 + rowOffset: 30,
	}
	top := excelize.Border{Type: "top", Style: 1, Color: "DADEE0"}
	left := excelize.Border{Type: "left", Style: 1, Color: "DADEE0"}
	right := excelize.Border{Type: "right", Style: 1, Color: "DADEE0"}
	bottom := excelize.Border{Type: "bottom", Style: 1, Color: "DADEE0"}

	// set each cell value
	for r, row := range data {
		if addr, err = excelize.JoinCellName(string('B'+colOffset), r); err != nil {
			fmt.Println(err)
			return err
		}
		if err = f.SetSheetRow(sheet, addr, &row); err != nil {
			fmt.Println(err)
			return err
		}
	}
	// set custom row height
	for r, ht := range height {
		if err = f.SetRowHeight(sheet, r, ht); err != nil {
			fmt.Println(err)
			return err
		}
	}

	// set custom column width
	if err = f.SetColWidth(sheet, string('B'+colOffset), string('H'+colOffset), 15); err != nil {
		fmt.Println(err)
		return err
	}

	// merge cell for the 'MONTH'
	if err = f.MergeCell(sheet, fmt.Sprintf("%s%d", string('B'+colOffset), 1+rowOffset), fmt.Sprintf("%s%d", string('D'+colOffset), 1+rowOffset)); err != nil {
		fmt.Println(err)
		return err
	}

	// define font style for the 'MONTH'
	if monthStyle, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "1f7f3b", Bold: true, Size: 22, Family: "Arial"},
	}); err != nil {
		fmt.Println(err)
		return err
	}

	// set font style for the 'MONTH'
	if err = f.SetCellStyle(sheet, fmt.Sprintf("%s%d", string('B'+colOffset), 1+rowOffset), fmt.Sprintf("%s%d", string('D'+colOffset), 1+rowOffset), monthStyle); err != nil {
		fmt.Println(err)
		return err
	}

	// define style for the 'SUNDAY' to 'SATURDAY'
	if titleStyle, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "1f7f3b", Size: 10, Bold: true, Family: "Arial"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E6F4EA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center"},
		Border:    []excelize.Border{{Type: "top", Style: 2, Color: "1f7f3b"}},
	}); err != nil {
		fmt.Println(err)
		return err
	}

	// set style for the 'SUNDAY' to 'SATURDAY'
	if err = f.SetCellStyle(sheet, fmt.Sprintf("%s%d", string('B'+colOffset), 3+rowOffset), fmt.Sprintf("%s%d", string('H'+colOffset), 3+rowOffset), titleStyle); err != nil {
		fmt.Println(err)
		return err
	}

	// define cell border for the date cell in the date range
	if dataStyle, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{top, left, right},
	}); err != nil {
		fmt.Println(err)
		return err
	}

	// set cell border for the date cell in the date range
	for _, r := range []int{4, 6, 8, 10, 12, 14} {
		if err = f.SetCellStyle(sheet, fmt.Sprintf("%s%d", string('B'+colOffset), r+rowOffset),
			fmt.Sprintf("%s%d", string('H'+colOffset), r+rowOffset), dataStyle); err != nil {
			fmt.Println(err)
			return err
		}
	}

	// define cell border for the blank cell in the date range
	if blankStyle, err = f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{left, right, bottom},
		Font:      &excelize.Font{Size: 9},
		Alignment: &excelize.Alignment{WrapText: true},
	}); err != nil {
		fmt.Println(err)
		return err
	}

	// set cell border for the blank cell in the date range, but only for cells that don't have a fill color
	for _, r := range []int{5, 7, 9, 11, 13, 15} {
		for c := 0; c < 7; c++ {
			cellAddr, _ := excelize.JoinCellName(string('B'+colOffset+c), r+rowOffset)
			if styleID, err := f.GetCellStyle(sheet, cellAddr); err != nil {
				// Handle error
				fmt.Println("Error mendapatkan gaya sel:", err)
				return err
			} else if styleID == 0 {
				// Jika tidak ada gaya yang diterapkan, maka terapkan blankStyle
				if err = f.SetCellStyle(sheet, cellAddr, cellAddr, blankStyle); err != nil {
					fmt.Println("Error menerapkan blankStyle:", err)
					return err
				}
			}
		}
	}

	// hide gridlines for the worksheet
	disable := false
	if err := f.SetSheetView(sheet, 0, &excelize.ViewOptions{
		ShowGridLines: &disable,
	}); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func ExportAll(c *gin.Context) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	f.DeleteSheet("Sheet1")

	sheets := []struct {
		name     string
		exporter func(c *gin.Context, f *excelize.File, sheetName string, isStandAlone bool) error
	}{
		{"BOOKING RAPAT", ExportBookingRapatToExcel},
		{"TIMELINE PROJECT", ExportTimelineProjectToExcel},
		{"TIMELINE DESKTOP", ExportTimelineDesktopToExcel},
		{"BERITA ACARA", ExportBeritaAcaraToExcel},
		{"PROJECT", ExportProjectToExcel},
		{"MEETING", ExportMeetingToExcel},
	}

	for _, sheet := range sheets {
		if err := sheet.exporter(c, f, sheet.name, false); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Error exporting %s: %v", sheet.name, err)})
			return
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error writing to buffer: %v", err)})
		return
	}

	fileName := "its_report_all.xlsx"
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
