package db

import (
	LogController "../controllers/logg"
	"../models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB
var err error

// Init creates a connection to mysql database and
// migrates any new models
func Init() {
	db, err = gorm.Open("mysql", "root:password@/b?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		LogController.Error("Failed to connect to database")
		panic(err)
	}
	LogController.Info("Database connected")

	if !db.HasTable(&models.User{}) {
		err := db.CreateTable(&models.User{})
		if err != nil {
			LogController.Error("Table already exists")
		}
	}

	if !db.HasTable(&models.Bank{}) {
		err := db.CreateTable(&models.Bank{})
		if err != nil {
			LogController.Error("Table already exists")
		}
	}

	if !db.HasTable(&models.Queue{}) {
		err := db.CreateTable(&models.Queue{})
		if err != nil {
			LogController.Error("Table already exists")
		}
	}

	db.AutoMigrate(&models.User{}, &models.Bank{}, &models.Queue{})

}

//GetDB ...
func GetDB() *gorm.DB {
	return db
}

func CloseDB() {
	db.Close()
}
