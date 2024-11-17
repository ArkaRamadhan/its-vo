package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	helper "github.com/arkaramadhan/its-vo/common/utils"

)

// ************* Model and Config Memo ************* //
type Memo struct {
	ID        uint       `gorm:"primaryKey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	Tanggal   *time.Time `json:"tanggal"`
	NoMemo    *string    `json:"no_memo"`
	Perihal   *string    `json:"perihal"`
	Pic       *string    `json:"pic"`
	CreateBy  string     `json:"create_by"`
}

// MarshalJSON menyesuaikan serialisasi JSON untuk struct Memo
func (i *Memo) MarshalJSON() ([]byte, error) {
	type Alias Memo
	if i.Tanggal == nil {
		return json.Marshal(&struct {
			Tanggal string `json:"tanggal"`
			*Alias
		}{
			Tanggal: "", // Atau nilai default lain yang Anda inginkan
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

func (ba *Memo) ToExcelRow() []interface{} {
	tanggalStr := ""
	if ba.Tanggal != nil {
		tanggalStr = ba.Tanggal.Format("2006-01-02")
	}

	return []interface{}{
		tanggalStr,
		helper.GetValue(ba.NoMemo),
		helper.GetValue(ba.Perihal),
		helper.GetValue(ba.Pic),
	}
}

func (ba *Memo) GetDocType() string {
	noMemo := helper.GetValue(ba.NoMemo)
	if strings.Contains(noMemo, "ITS-SAG") {
		return "SAG"
	}
	return "ISO"
}

func (m *Memo) SetProperty(key string, value interface{}) error {
	switch key {
	case "Tanggal":
		val, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("type assertion to time.Time failed for Tanggal")
		}
		m.Tanggal = &val
	case "NoMemo":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for NoMemo")
		}
		m.NoMemo = &val
	case "Perihal":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Perihal")
		}
		m.Perihal = &val
	case "Pic":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Pic")
		}
		m.Pic = &val
	default:
		return fmt.Errorf("unknown property: %s", key)
	}
	return nil
}

// ************* Model and Config Berita Acara ************* //
type BeritaAcara struct {
	ID        uint       `gorm:"primaryKey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	NoSurat   *string    `json:"no_surat"`
	Tanggal   *time.Time `json:"tanggal"`
	Perihal   *string    `json:"perihal"`
	Pic       *string    `json:"pic"`
	CreateBy  string     `json:"create_by"`
}

func (i *BeritaAcara) MarshalJSON() ([]byte, error) {
	type Alias BeritaAcara
	if i.Tanggal == nil {
		return json.Marshal(&struct {
			Tanggal string `json:"tanggal"`
			*Alias
		}{
			Tanggal: "", // Atau nilai default lain yang Anda inginkan
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

func (ba *BeritaAcara) ToExcelRow() []interface{} {
	tanggalStr := ""
	if ba.Tanggal != nil {
		tanggalStr = ba.Tanggal.Format("2006-01-02")
	}

	return []interface{}{
		tanggalStr,
		helper.GetValue(ba.NoSurat),
		helper.GetValue(ba.Perihal),
		helper.GetValue(ba.Pic),
	}
}

func (ba *BeritaAcara) GetDocType() string {
	noSurat := helper.GetValue(ba.NoSurat)
	if strings.Contains(noSurat, "ITS-SAG") {
		return "SAG"
	}
	return "ISO"
}

func (ba *BeritaAcara) SetProperty(key string, value interface{}) error {
	switch key {
	case "Tanggal":
		val, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("type assertion to time.Time failed for Tanggal")
		}
		ba.Tanggal = &val
	case "NoSurat":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for NoSurat")
		}
		ba.NoSurat = &val
	case "Perihal":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Perihal")
		}
		ba.Perihal = &val
	case "Pic":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Pic")
		}
		ba.Pic = &val
	default:
		return fmt.Errorf("unknown property: %s", key)
	}
	return nil
}

// ************* Model and Config Surat ************* //
type Surat struct {
	ID        uint       `gorm:"primaryKey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	NoSurat   *string    `json:"no_surat"`
	Tanggal   *time.Time `json:"tanggal"`
	Perihal   *string    `json:"perihal"`
	Pic       *string    `json:"pic"`
	CreateBy  string     `json:"create_by"`
}

func (i *Surat) MarshalJSON() ([]byte, error) {
	type Alias Surat
	if i.Tanggal == nil {
		return json.Marshal(&struct {
			Tanggal string `json:"tanggal"`
			*Alias
		}{
			Tanggal: "", // Atau nilai default lain yang Anda inginkan
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

func (s *Surat) ToExcelRow() []interface{} {
	tanggalStr := ""
	if s.Tanggal != nil {
		tanggalStr = s.Tanggal.Format("2006-01-02")
	}

	return []interface{}{
		tanggalStr,
		helper.GetValue(s.NoSurat),
		helper.GetValue(s.Perihal),
		helper.GetValue(s.Pic),
	}
}

func (s *Surat) GetDocType() string {
	noSurat := helper.GetValue(s.NoSurat)
	if strings.Contains(noSurat, "ITS-SAG") {
		return "SAG"
	}
	return "ISO"
}

func (s *Surat) SetProperty(key string, value interface{}) error {
	switch key {
	case "Tanggal":
		val, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("type assertion to time.Time failed for Tanggal")
		}
		s.Tanggal = &val
	case "NoSurat":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for NoSurat")
		}
		s.NoSurat = &val
	case "Perihal":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Perihal")
		}
		s.Perihal = &val
	case "Pic":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Pic")
		}
		s.Pic = &val
	default:
		return fmt.Errorf("unknown property: %s", key)
	}
	return nil
}

// ************* Model and Config Sk ************* //
type Sk struct {
	ID        uint       `gorm:"primaryKey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	NoSurat   *string    `json:"no_surat"`
	Tanggal   *time.Time `json:"tanggal"`
	Perihal   *string    `json:"perihal"`
	Pic       *string    `json:"pic"`
	CreateBy  string     `json:"create_by"`
}

func (i *Sk) MarshalJSON() ([]byte, error) {
	type Alias Sk
	if i.Tanggal == nil {
		return json.Marshal(&struct {
			Tanggal string `json:"tanggal"`
			*Alias
		}{
			Tanggal: "", // Atau nilai default lain yang Anda inginkan
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

func (ba *Sk) ToExcelRow() []interface{} {
	tanggalStr := ""
	if ba.Tanggal != nil {
		tanggalStr = ba.Tanggal.Format("2006-01-02")
	}

	return []interface{}{
		tanggalStr,
		helper.GetValue(ba.NoSurat),
		helper.GetValue(ba.Perihal),
		helper.GetValue(ba.Pic),
	}
}

func (ba *Sk) GetDocType() string {
	noSurat := helper.GetValue(ba.NoSurat)
	if strings.Contains(noSurat, "ITS-SAG") {
		return "SAG"
	}
	return "ISO"
}

func (s *Sk) SetProperty(key string, value interface{}) error {
	switch key {
	case "Tanggal":
		val, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("type assertion to time.Time failed for Tanggal")
		}
		s.Tanggal = &val
	case "NoSurat":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for NoSurat")
		}
		s.NoSurat = &val
	case "Perihal":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Perihal")
		}
		s.Perihal = &val
	case "Pic":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Pic")
		}
		s.Pic = &val
	default:
		return fmt.Errorf("unknown property: %s", key)
	}
	return nil
}

type Perdin struct {
	ID        uint       `gorm:"primaryKey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	NoPerdin  *string    `json:"no_perdin"`
	Tanggal   *time.Time `json:"tanggal"`
	Hotel     *string    `json:"hotel"`
	Transport *string    `json:"transport"`
	CreateBy  string     `json:"create_by"`
}

func (i *Perdin) MarshalJSON() ([]byte, error) {
	type Alias Perdin
	if i.Tanggal == nil {
		return json.Marshal(&struct {
			Tanggal string `json:"tanggal"`
			*Alias
		}{
			Tanggal: "", // Atau nilai default lain yang Anda inginkan
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

func (p *Perdin) ToExcelRow() []interface{} {
	tanggalStr := ""
	if p.Tanggal != nil {
		tanggalStr = p.Tanggal.Format("2006-01-02")
	}

	return []interface{}{
		helper.GetValue(p.NoPerdin),
		tanggalStr,
		helper.GetValue(p.Hotel),
		helper.GetValue(p.Transport),
	}
}

func (p *Perdin) GetDocType() string {
	noPerdin := helper.GetValue(p.NoPerdin)
	if strings.Contains(noPerdin, "ITS-SAG") {
		return "SAG"
	}
	return "ISO"
}

func (p *Perdin) SetProperty(key string, value interface{}) error {
	switch key {
	case "Tanggal":
		val, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("type assertion to time.Time failed for Tanggal")
		}
		p.Tanggal = &val
	case "NoPerdin":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for NoPerdin")
		}
		p.NoPerdin = &val
	case "Hotel":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Hotel")
		}
		p.Hotel = &val
	case "Transport":
		val, ok := value.(string)
		if !ok {
			return fmt.Errorf("type assertion to string failed for Transport")
		}
		p.Transport = &val
	default:
		return fmt.Errorf("unknown property: %s", key)
	}
	return nil
}
