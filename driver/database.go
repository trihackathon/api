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

	if err := db.AutoMigrate(
		&models.User{},
		&models.Activity{},
		&models.GPSPoint{},
		&models.Team{},
		&models.TeamMember{},
		&models.InviteCode{},
		&models.Goal{},
		&models.WeeklyEvaluation{},
		&models.DisbandVote{},
		&models.ActivityReview{},
	); err != nil {
		log.Fatalf("マイグレーションエラー: %v", err)
	}

	log.Println("DB接続成功")
	return db
}
