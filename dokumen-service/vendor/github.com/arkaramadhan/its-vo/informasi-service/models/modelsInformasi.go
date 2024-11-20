package models

import (
	"encoding/json"
	"strings"
	"time"

	helper "github.com/arkaramadhan/its-vo/common/utils"
	"gorm.io/gorm"

)

type SuratMasuk struct {
	ID         uint       `gorm:"primaryKey"`
	CreatedAt  *time.Time `gorm:"autoCreateTime"`
	UpdatedAt  *time.Time `gorm:"autoUpdateTime"`
	NoSurat    *string    `json:"no_surat"`
	Title      *string    `json:"title"`
	RelatedDiv *string    `json:"related_div"`
	DestinyDiv *string    `json:"destiny_div"`
	Tanggal    *time.Time `json:"tanggal"`
	CreateBy   string     `json:"create_by"`
}

func (i *SuratMasuk) MarshalJSON() ([]byte, error) {
	type Alias SuratMasuk
	if i.Tanggal == nil {
		// Handle jika Tanggal nil
		return json.Marshal(&struct {
			Tanggal string `json:"tanggal"`
			*Alias
		}{
			Tanggal: "", // Atau format default yang diinginkan
			Alias:   (*Alias)(i),
		})
	} else {
		tanggalFormatted := i.Tanggal.Format("2006-01-02")
		return json.Marshal(&struct {
			Tanggal string `json:"tanggal"`
			*Alias
		}{
			Tanggal: tanggalFormatted,
			Alias:   (*Alias)(i),
		})
	}
}

func (ba *SuratMasuk) ToExcelRow() []interface{} {
	tanggalStr := ""
	if ba.Tanggal != nil {
		tanggalStr = ba.Tanggal.Format("2006-01-02")
	}

	return []interface{}{
		tanggalStr,
		helper.GetValue(ba.NoSurat),
		helper.GetValue(ba.Title),
		helper.GetValue(ba.RelatedDiv),
		helper.GetValue(ba.DestinyDiv),
	}
}

func (ba *SuratMasuk) GetDocType() string {
	noSurat := helper.GetValue(ba.NoSurat)
	if strings.Contains(noSurat, "ITS-SAG") {
		return "SAG"
	}
	return "ISO"
}

// model for suratKeluar
type SuratKeluar struct {
	ID        uint       `gorm:"primaryKey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	NoSurat   *string    `json:"no_surat"`
	Title     *string    `json:"title"`
	From      *string    `json:"from"`
	Pic       *string    `json:"pic"`
	Tanggal   *time.Time `json:"tanggal"`
	CreateBy  string     `json:"create_by"`
}

func (i *SuratKeluar) MarshalJSON() ([]byte, error) {
	type Alias SuratKeluar
	tanggalFormatted := i.Tanggal.Format("2006-01-02")
	return json.Marshal(&struct {
		Tanggal *string `json:"tanggal"`
		*Alias
	}{
		Tanggal: &tanggalFormatted,
		Alias:   (*Alias)(i),
	})
}

func (ba *SuratKeluar) ToExcelRow() []interface{} {
	tanggalStr := ""
	if ba.Tanggal != nil {
		tanggalStr = ba.Tanggal.Format("2006-01-02")
	}

	return []interface{}{
		tanggalStr,
		helper.GetValue(ba.NoSurat),
		helper.GetValue(ba.Title),
		helper.GetValue(ba.From),
		helper.GetValue(ba.Pic),
	}
}

func (ba *SuratKeluar) GetDocType() string {
	noSurat := helper.GetValue(ba.NoSurat)
	if strings.Contains(noSurat, "ITS-SAG") {
		return "SAG"
	}
	return "ISO"
}

type Arsip struct {
	gorm.Model
	NoArsip           *string    `json:"no_arsip"`
	JenisDokumen      *string    `json:"jenis_dokumen"`
	NoDokumen         *string    `json:"no_dokumen"`
	TanggalDokumen    *time.Time `json:"tanggal_dokumen"`
	Perihal           *string    `json:"perihal"`
	NoBox             *string    `json:"no_box"`
	TanggalPenyerahan *time.Time `json:"tanggal_penyerahan"`
	Keterangan        *string    `json:"keterangan"`
	CreateBy          string     `json:"create_by"`
}

func (a *Arsip) MarshalJSON() ([]byte, error) {
	type Alias Arsip
	var tanggalDokumenFormatted, tanggalPenyerahanFormatted string

	// Cek TanggalDokumen
	if a.TanggalDokumen == nil {
		tanggalDokumenFormatted = ""
	} else {
		tanggalDokumenFormatted = a.TanggalDokumen.Format("2006-01-02")
	}

	// Cek TanggalPenyerahan
	if a.TanggalPenyerahan == nil {
		tanggalPenyerahanFormatted = ""
	} else {
		tanggalPenyerahanFormatted = a.TanggalPenyerahan.Format("2006-01-02")
	}

	return json.Marshal(&struct {
		TanggalDokumen    string `json:"tanggal_dokumen"`
		TanggalPenyerahan string `json:"tanggal_penyerahan"`
		*Alias
	}{
		TanggalDokumen:    tanggalDokumenFormatted,
		TanggalPenyerahan: tanggalPenyerahanFormatted,
		Alias:             (*Alias)(a),
	})
}

func (ba *Arsip) ToExcelRow() []interface{} {
	tanggalStr := ""
	if ba.TanggalDokumen != nil {
		tanggalStr = ba.TanggalDokumen.Format("2006-01-02")
	}

	tanggalPenyerahanStr := ""
	if ba.TanggalPenyerahan != nil {
		tanggalPenyerahanStr = ba.TanggalPenyerahan.Format("2006-01-02")
	}

	return []interface{}{
		helper.GetValue(ba.NoArsip),
		helper.GetValue(ba.JenisDokumen),
		helper.GetValue(ba.NoDokumen),
		helper.GetValue(ba.Perihal),
		helper.GetValue(ba.NoBox),
		tanggalStr,
		tanggalPenyerahanStr,
		helper.GetValue(ba.Keterangan),
	}
}

func (ba *Arsip) GetDocType() string {
	noArsip := helper.GetValue(ba.NoArsip)
	if strings.Contains(noArsip, "ITS-SAG") {
		return "SAG"
	}
	return "ISO"
}
