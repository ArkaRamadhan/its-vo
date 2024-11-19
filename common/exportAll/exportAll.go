package exportAll

import (
	"bytes"
	"fmt"

	// informasi "github.com/arkaramadhan/its-vo/informasi-service/controllers"
	// weekly "github.com/arkaramadhan/its-vo/weeklyTimeline-service/controllers"
	// project "github.com/arkaramadhan/its-vo/project-service/controllers"
	dokumen "github.com/arkaramadhan/its-vo/dokumen-service/controllers"
	// kegiatan "github.com/arkaramadhan/its-vo/kegiatan-service/controllers"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

)

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
		{"BERITA ACARA", dokumen.ExportBeritaAcaraToExcel},
		{"MEMO", dokumen.ExportMemoToExcel},
		{"PERDIN", dokumen.ExportPerdinToExcel},
		{"SK", dokumen.ExportSkToExcel},
		{"SURAT", dokumen.ExportSuratToExcel},

		{"ARSIP", informasi.ExportArsipToExcel},
		{"SURAT KELUAR", informasi.ExportSuratKeluarToExcel},
		{"SURAT MASUK", informasi.ExportSuratMasukToExcel},

		{"BOOKING RAPAT", kegiatan.ExportBookingRapatToExcel},
		{"JADWAL CUTI", kegiatan.ExportJadwalCutiToExcel},
		{"JADWAL RAPAT", kegiatan.ExportJadwalRapatToExcel},
		{"MEETING", kegiatan.ExportMeetingToExcel},
		{"TIMELINE DESKTOP", kegiatan.ExportTimelineDesktopToExcel},

		{"PROJECT", project.ExportProjectToExcel},

		{"TIMELINE PROJECT", weekly.ExportTimelineProjectToExcel},
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
