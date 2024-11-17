package main

import (
	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/informasi-service/models"
)

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDB("informasi")
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

}

func main() {

	initializers.DB.AutoMigrate(
		&models.Arsip{},
		&models.SuratMasuk{},
		&models.SuratKeluar{},
	)

}
