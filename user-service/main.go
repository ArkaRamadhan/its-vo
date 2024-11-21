package main

import (
	"log"
	"os"

	"github.com/arkaramadhan/its-vo/common/initializers"
	"github.com/arkaramadhan/its-vo/common/middleware"
	"github.com/arkaramadhan/its-vo/user-service/controllers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

)

func init() {
	initializers.LoadEnvVariables()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set in the environment variables")
	}

	err := initializers.ConnectToDB("user")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func main() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(middleware.CORS())

	r.POST("/login", controllers.Login)

	r.Use(middleware.TokenAuthMiddleware())

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/user", controllers.UserIndex)
	r.POST("/user", controllers.Register)
	r.DELETE("/user/:id", controllers.UserDelete)
	r.PUT("/user/:id", controllers.UserUpdate)

	r.POST("/logout", controllers.Logout)

	r.Run(":8084")
}
