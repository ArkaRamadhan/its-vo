package main

import (
	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/user-service/models"
)

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDB("user")
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

}

func main() {

	initializers.DB.AutoMigrate(
		&models.User{},
		&models.UserToken{},
	)

}
