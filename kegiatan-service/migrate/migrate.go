package main

import (
	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/kegiatan-service/models"
)

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDB("kegiatan")
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

}

func main() {

	initializers.DB.AutoMigrate(
		&models.TimelineDesktop{},
		&models.ResourceDesktop{},
		&models.BookingRapat{},
		&models.JadwalRapat{},
		&models.JadwalCuti{},
		&models.ConflictRequest{},
		&models.Notification{},
		&models.Meeting{},
	)

}
