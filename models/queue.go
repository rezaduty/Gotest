package models

type Queue struct {
	ID                      uint `gorm:"primary_key;AUTO_INCREMENT"`
	Bank                    *Bank
	BankID                  int `json:"-"`
	User                    *User
	UserID                  int `json:"-"`
	Turn_number             int `gorm:"default:'0'"`
	Turn_status             int `gorm:"default:'0'"`
	Temp_Active_Turn_Number int `gorm:"default:'0'"`
	Method                  int `gorm:"default:'0'"`
	Cancel                  bool
}
