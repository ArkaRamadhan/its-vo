package main

import (
	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/dokumen-service/models"
)

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDB("dokumen")
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

}

func main() {

	initializers.DB.AutoMigrate(
		&models.Memo{},
		&models.Surat{},
		&models.Perdin{},
		&models.Sk{},
		&models.BeritaAcara{},
	)

}
