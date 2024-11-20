package models

import "time"

type File struct {
	ID          uint      `gorm:"primaryKey"`     // ID unik untuk file
	CreatedAt   time.Time `gorm:"autoCreateTime"` // Timestamp saat file diunggah
	UpdatedAt   time.Time `gorm:"autoUpdateTime"` // Timestamp untuk setiap update
	UserID      uint      `gorm:"index"`          // ID pengguna yang mengunggah file
	FilePath    string    `gorm:"not null"`       // Path lengkap di mana file disimpan
	FileName    string    `gorm:"not null"`       // Nama file asli
	ContentType string    `gorm:"not null"`       // Jenis konten file, misal 'application/pdf'
	Size        int64     `gorm:"not null"`       // Ukuran file dalam byte
}

func (File) Tablefiles() string {
    return "common.files"  // Menentukan schema.table_name
}

type Notification struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Title    string    `json:"title"`
	Start    time.Time `json:"start"`
	Category string    `json:"category"`
}

func (File) TableNotification() string {
    return "common.notifications"  // Menentukan schema.table_name
}
