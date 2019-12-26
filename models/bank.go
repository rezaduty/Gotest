package models

type Bank struct {
	ID uint `json:"bank_id";gorm:"primary_key;AUTO_INCREMENT"`

	Users []*User

	Name                string `json:"name"  binding:"required"`
	Branch              int    `json:"branch"`
	Code                int    `json:"code"`
	Address             string `json:"address"`
	X                   string
	Y                   string
	Turns_number        int `gorm:"default:'0'"`
	Active_turns_number int `gorm:"default:'0'"`
}
