package main

import (
	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/models"
)

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDB("common")
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

}

func main() {

	initializers.DB.AutoMigrate(
		&models.File{},
		&models.Notification{},
	)

}
