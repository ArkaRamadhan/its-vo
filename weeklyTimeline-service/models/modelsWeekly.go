package models

import (
	"encoding/json"
	"time"
)

type TimelineProject struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Start      string `json:"start"`
	End        string `json:"end"`
	ResourceId int    `json:"resourceId"` // Ubah tipe data dari string ke int
	Title      string `json:"title"`
	BgColor    string `json:"bgColor"`
}

func (TimelineProject) TableName() string {
	return "timeline_projects"
}

type ResourceProject struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	ParentID uint   `json:"parent_id"`
}

func (e TimelineProject) GetTitle() string {
	return e.Title
}

func (e TimelineProject) GetStart() time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", e.Start) // Tambahkan penanganan error yang sesuai
	return t
}

func (e TimelineProject) GetEnd() time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", e.End) // Tambahkan penanganan error yang sesuai
	return t
}

func (e TimelineProject) GetColor() string {
	return e.BgColor
}

func (e TimelineProject) GetResourceID() uint {
	return uint(e.ResourceId)
}

func (e TimelineProject) GetAllDay() bool {
	return false
}

type MeetingSchedule struct {
	ID        uint       `gorm:"primaryKey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	Hari      *string    `json:"hari"`
	Tanggal   *time.Time `json:"tanggal"`
	Perihal   *string    `json:"perihal"`
	Waktu     *string    `json:"waktu"`
	Selesai   *string    `json:"selesai"`
	Tempat    *string    `json:"tempat"`
	Pic       *string    `json:"pic"`
	Status    *string    `json:"status"`
	CreateBy  string     `json:"create_by"`
	Color     string     `json:"color"`
}

func (i *MeetingSchedule) MarshalJSON() ([]byte, error) {
	type Alias MeetingSchedule
	tanggalFormatted := i.Tanggal.Format("2006-01-02")
	return json.Marshal(&struct {
		Tanggal *string `json:"tanggal"`
		*Alias
	}{
		Tanggal: &tanggalFormatted,
		Alias:   (*Alias)(i),
	})
}
