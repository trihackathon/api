package driver

import (
	"log"
	"os"

	"github.com/trihackathon/api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB接続エラー: %v", err)
	}

	// AutoMigrate
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("マイグレーションエラー: %v", err)
	}

	log.Println("DB接続成功")
	return db
}
