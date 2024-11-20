package initializers

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB(schema string) error {
	var err error
	baseDSN := os.Getenv("DB_URL")
	if baseDSN == "" {
		log.Println("DB_URL is not set in the environment variables")
		return fmt.Errorf("DB_URL is not set in the environment variables")
	}

	dsn := fmt.Sprintf("%s search_path=%s", baseDSN, schema)
	log.Printf("Connecting to database with DSN: %s\n", dsn)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		log.Printf("Failed to connect to database with schema %s: %v\n", schema, err)
		return err
	}

	log.Println("Successfully connected to the database")
	log.Printf("DB Connection Details: %+v\n", DB)
	return nil
}

func main() {
	ConnectToDB("your_schema_here")
}
