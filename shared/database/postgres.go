package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectSupabase() *gorm.DB {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		os.Getenv("SUPABASE_HOST"),
		os.Getenv("SUPABASE_USER"),
		os.Getenv("SUPABASE_PASSWORD"),
		os.Getenv("SUPABASE_DB"),
		os.Getenv("SUPABASE_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect Supabase: %v", err)
	}

	log.Println("✅ Connected to Supabase")

	return db
}
