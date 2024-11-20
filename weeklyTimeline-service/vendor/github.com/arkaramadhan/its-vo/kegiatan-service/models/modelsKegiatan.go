package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	helper "github.com/arkaramadhan/its-vo/common/utils"

)

type Meeting struct {
	ID               uint       `gorm:"primaryKey"`
	CreatedAt        *time.Time `gorm:"autoCreateTime"`
	UpdatedAt        *time.Time `gorm:"autoUpdateTime"`
	Task             *string    `json:"task"`
	TindakLanjut     *string    `json:"tindak_lanjut"`
	Status           *string    `json:"status"`
	UpdatePengerjaan *string    `json:"update_pengerjaan"`
	Pic              *string    `json:"pic"`
	TanggalTarget    *time.Time `json:"tanggal_target"`
	TanggalActual    *time.Time `json:"tanggal_actual"`
	CreateBy         string     `json:"create_by"`
}

func (i *Meeting) MarshalJSON() ([]byte, error) {
	type Alias Meeting
	tanggalTargetFormatted := i.TanggalTarget.Format("2006-01-02")
	tanggalActualFormatted := i.TanggalActual.Format("2006-01-02")
	return json.Marshal(&struct {
		TanggalTarget *string `json:"tanggal_target"`
		TanggalActual *string `json:"tanggal_actual"`
		*Alias
	}{
		TanggalTarget: &tanggalTargetFormatted,
		TanggalActual: &tanggalActualFormatted,
		Alias:         (*Alias)(i),
	})
}

func (mt *Meeting) ToExcelRow() []interface{} {
	tanggalTargetStr := ""
	tanggalActualStr := ""
	if mt.TanggalTarget != nil {
		tanggalTargetStr = mt.TanggalTarget.Format("2006-01-02")
	}
	if mt.TanggalActual != nil {
		tanggalActualStr = mt.TanggalActual.Format("2006-01-02")
	}

	return []interface{}{
		helper.GetValue(mt.Task),
		helper.GetValue(mt.TindakLanjut),
		helper.GetValue(mt.Status),
		helper.GetValue(mt.UpdatePengerjaan),
		helper.GetValue(mt.Pic),
		tanggalTargetStr,
		tanggalActualStr,
	}
}

func (mt *Meeting) GetDocType() string {
	return "MEETING"
}

func (m *Meeting) GetStatus() string {
	if m.Status == nil {
		return "" // atau nilai default lainnya
	}
	return *m.Status
}

type BookingRapat struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Title  string `json:"title"`
	Start  string `json:"start"`
	End    string `json:"end"`
	AllDay bool   `json:"allDay"`
	Color  string `json:"color"` // Tambahkan field ini untuk warna
	Status string `json:"status"`
}

func (BookingRapat) TableName() string {
	return "booking_rapats"
}

func (e BookingRapat) GetTitle() string {
	return e.Title
}

func (e BookingRapat) GetStart() time.Time {
	if e.AllDay {
		t, _ := time.Parse("2006-01-02", e.Start)
		return t
	} else {
		t, _ := time.Parse(time.RFC3339, e.Start) // Tambahkan penanganan error yang sesuai
		return t
	}
}

func (e BookingRapat) GetEnd() time.Time {
	if e.AllDay {
		t, _ := time.Parse("2006-01-02", e.End)
		return t
	} else {
		t, _ := time.Parse(time.RFC3339, e.End) // Tambahkan penanganan error yang sesuai
		return t
	}
}

func (e BookingRapat) GetColor() string {
	return e.Color
}

func (e BookingRapat) GetAllDay() bool {
	return e.AllDay
}

func (e BookingRapat) GetResourceID() uint {
	return 0
}

type JadwalRapat struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Title  string `json:"title"`
	Start  string `json:"start"`
	End    string `json:"end"`
	AllDay bool   `json:"allDay"`
	Color  string `json:"color"`
}

func (JadwalRapat) TableName() string {
	return "jadwal_rapats"
}

func (e JadwalRapat) GetTitle() string {
	return e.Title
}

func (e JadwalRapat) GetStart() time.Time {
	if e.AllDay {
		t, _ := time.Parse("2006-01-02", e.Start)
		return t
	} else {
		t, _ := time.Parse(time.RFC3339, e.Start) // Tambahkan penanganan error yang sesuai
		return t
	}
}

func (e JadwalRapat) GetEnd() time.Time {
	if e.AllDay {
		t, _ := time.Parse("2006-01-02", e.End)
		return t
	} else {
		t, _ := time.Parse(time.RFC3339, e.End) // Tambahkan penanganan error yang sesuai
		return t
	}
}

func (e JadwalRapat) GetColor() string {
	return e.Color
}

func (e JadwalRapat) GetAllDay() bool {
	return e.AllDay
}

func (e JadwalRapat) GetResourceID() uint {
	return 0
}

type JadwalCuti struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Title  string `json:"title"`
	Start  string `json:"start"`
	End    string `json:"end"`
	AllDay bool   `json:"allDay"`
	Color  string `json:"color"` // Tambahkan field ini untuk warna
}

func (e JadwalCuti) TableName() string {
	return "jadwal_cutis"
}

func (e JadwalCuti) GetTitle() string {
	return e.Title
}

func (e JadwalCuti) GetStart() time.Time {
	if e.AllDay {
		t, _ := time.Parse("2006-01-02", e.Start)
		return t
	} else {
		t, _ := time.Parse(time.RFC3339, e.Start) // Tambahkan penanganan error yang sesuai
		return t
	}
}

func (e JadwalCuti) GetEnd() time.Time {
	if e.AllDay {
		t, _ := time.Parse("2006-01-02", e.End)
		return t
	} else {
		t, _ := time.Parse(time.RFC3339, e.End) // Tambahkan penanganan error yang sesuai
		return t
	}
}

func (e JadwalCuti) GetColor() string {
	return e.Color
}

func (e JadwalCuti) GetAllDay() bool {
	return e.AllDay
}

func (e JadwalCuti) GetResourceID() uint {
	return 0
}

type TimelineDesktop struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Start      string `json:"start"`
	End        string `json:"end"`
	ResourceId int    `json:"resourceId"` // Ubah tipe data dari string ke int
	Title      string `json:"title"`
	BgColor    string `json:"bgColor"`
}

func (TimelineDesktop) TableName() string {
	return "timeline_desktops"
}

func (e TimelineDesktop) GetTitle() string {
	return e.Title
}

func (e TimelineDesktop) GetStart() time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", e.Start) // Tambahkan penanganan error yang sesuai
	return t
}

func (e TimelineDesktop) GetEnd() time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", e.End) // Tambahkan penanganan error yang sesuai
	return t
}

func (e TimelineDesktop) GetColor() string {
	return e.BgColor
}

func (e TimelineDesktop) GetResourceID() uint {
	return uint(e.ResourceId)
}

func (e TimelineDesktop) GetAllDay() bool {
	return false
}

type ResourceDesktop struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	ParentID uint   `json:"parent_id"`
}

type ConflictRequest struct {
	gorm.Model
	NewEventID uint
	OldEventID uint
	Status     string
	OldTitle   string
	NewTitle   string
	StartTime  string
	EndTime    string
	Date       time.Time
}

type Notification struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Title    string    `json:"title"`
	Start    time.Time `json:"start"`
	Category string    `json:"category"`
}
