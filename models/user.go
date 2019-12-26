package models

type User struct {
	Banks        []*Bank
	ID           uint   `gorm:"primary_key;AUTO_INCREMENT"`
	Name         string `json:name`
	Password     string `json:"-"`
	Email        string `json:"-"`
	MobilePhone  string `json:"-"`
	ActiveStatus int    `gorm:"default:'0'"`
	CC           int    `gorm:"default:'0'"`
	Role         string
}

type JwtToken struct {
	Token string `json:"token"`
}

type Exception struct {
	Message string `json:"message"`
}
