package main

import (
	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/weeklyTimeline-service/models"
)

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDB("weekly_timeline")
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

}

func main() {

	initializers.DB.AutoMigrate(
		&models.MeetingSchedule{},
		&models.ResourceProject{},
		&models.TimelineProject{},
	)

}
