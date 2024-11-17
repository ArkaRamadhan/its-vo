package main

import (
	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/project-service/models"
)

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDB("project")
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

}

func main() {

	initializers.DB.AutoMigrate(
		&models.Project{},
	)

}
