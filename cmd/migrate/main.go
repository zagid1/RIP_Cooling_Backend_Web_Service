package main

import (
	"RIP/internal/app/ds"
	"RIP/internal/app/dsn"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func main() {
	_ = godotenv.Load()

	// Отладочная информация
	fmt.Println("Current directory:", getCurrentDir())
	fmt.Println("DB_USER:", os.Getenv("DB_USER"))
	fmt.Println("DB_PORT:", os.Getenv("DB_PORT"))
	fmt.Println("DB_NAME:", os.Getenv("DB_NAME"))

	dsnString := dsn.FromEnv()
	fmt.Println("Final DSN:", dsnString)

	db, err := gorm.Open(postgres.Open(dsnString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	err = db.AutoMigrate(&ds.Component{}, &ds.Cooling{}, &ds.ComponentToCooling{}, &ds.Users{})
	if err != nil {
		panic("cant migrate db: " + err.Error())
	}

	fmt.Println("Migration completed successfully!")
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}
