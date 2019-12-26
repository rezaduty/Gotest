package users

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/badoux/checkmail"
	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"

	LogController "../../controllers/logg"
	"github.com/valyala/fasthttp"

	"crypto/sha256"

	"../../models"

	"../../db"
)

/*
use in doJSONWrite func
*/
var (
	strContentType     = []byte("Content-Type")
	strApplicationJSON = []byte("application/json")
)

/*
check user is validate and auth
*/
func ValidateMiddleware(t string) string {
	ptoken := t
	token, _ := jwt.Parse(string(ptoken), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte("secret"), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user models.User
		mapstructure.Decode(claims, &user)

		return claims["username"].(string)
	} else {
		return string("0")
	}

}

/*
use in register func
*/
func isUsernameAlreadyExist(name string) bool {
	db := db.GetDB()
	user := &models.User{}

	db.Find(&user)
	if user.Name == name {
		return true
	} else {
		return false
	}
}

/*
use in register func
*/
func isEmailAlreadyExist(email string) bool {
	db := db.GetDB()
	user := &models.User{}

	db.Find(&user)
	if user.Email == email {
		return true
	} else {
		return false
	}
}

/*
use in register func
*/
func isPhoneAlreadyExist(phone string) bool {
	db := db.GetDB()
	user := &models.User{}

	db.Find(&user)
	if user.MobilePhone == phone {
		return true
	} else {
		return false
	}
}

/*
for return as json with data index
params
`code` for http status code
`jsonValue` for value of data index
*/
func doJSONWrite(ctx *fasthttp.RequestCtx, code int, jsonValue string) {
	var jsonStr = []byte(`{"data":"` + jsonValue + `"}`)
	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
	ctx.Response.SetStatusCode(code)
	ctx.Write([]byte(jsonStr))
}
func reverse(s string) string {
	cs := make([]rune, utf8.RuneCountInString(s))
	i := len(cs)
	for _, c := range s {
		i--
		cs[i] = c
	}
	return string(cs)
}

/*
register user with `name` and `password` and `email` and `MobilePhone`

url: http://localhost:8020/user/register
method:POST
body:
{
"Name":"ali",
"Password":"momy8",
"Email":"ali@gmail.com",
"MobilePhone":"989012122131"
}
*/
func Register(ctx *fasthttp.RequestCtx) {
	var UF = true
	var EF = true
	var PF = true
	db := db.GetDB()

	user := &models.User{}
	err := json.Unmarshal(ctx.PostBody(), &user)

	if err != nil {
		panic(err)
	}

	Name := user.Name
	s := user.Password
	h := sha256.New()
	h.Write([]byte(s))
	sha256_hash := hex.EncodeToString(h.Sum(nil))
	Password := sha256_hash
	Email := user.Email
	MobilePhone := user.MobilePhone
	ActiveStatus := 1
	Role := "client"
	ctx.Response.Header.Set("Content-Type", "application/json")
	if Name == "" || Password == "" || Email == "" {
		LogController.Error("Name Or Password Or Email null")
		ctx.SetBody([]byte("{data:401}"))
		return
	}

	err = checkmail.ValidateFormat(Email)
	if err != nil {

		doJSONWrite(ctx, 403, "Email Not Validate")
		LogController.Error("User Email Not Validate: " + Email)

	} else {
		if isUsernameAlreadyExist(Name) {
			UF = false
			doJSONWrite(ctx, 403, "Username Already")
			LogController.Warning("User Already Exist: " + Name)

		}
		if isEmailAlreadyExist(Email) {
			EF = false
			doJSONWrite(ctx, 403, "Email Already")
			LogController.Warning("User Already Exist: " + Email)
		}
		if isPhoneAlreadyExist(MobilePhone) {
			PF = false
			doJSONWrite(ctx, 403, "Phone Already")
			LogController.Warning("User Already Exist: " + MobilePhone)
		}

		if UF && EF && PF {
			var user = &models.User{Name: Name, Password: Password, Email: Email, MobilePhone: MobilePhone, ActiveStatus: ActiveStatus, Role: Role}
			db.Create(&user)

			doJSONWrite(ctx, 201, "")
			LogController.Notice("User Success Register: " + Email)

		}
	}

}

/*
if user auth with `phone` and `password` generate token
token like this
token = token+@+reverse(name)+!+3423423423423423jhb234g3h2ik53l5g34h5jgv234oy5gv43jh25v34l5v43uo25vy


url: http://localhost:8020/user/login
method:POST
body:
{
"MobilePhone":"0912121231",
"Password":"momy8"
}


*/
func Login(ctx *fasthttp.RequestCtx) {
	db := db.GetDB()

	user := &models.User{}
	err := json.Unmarshal(ctx.PostBody(), &user)

	if err != nil {
		panic(err)
	}
	s := user.Password
	h := sha256.New()
	h.Write([]byte(s))
	sha256_hash := hex.EncodeToString(h.Sum(nil))

	Phone := user.MobilePhone
	Password := sha256_hash

	db.Where(&models.User{MobilePhone: Phone, Password: Password}).Find(&user)

	if user.Name != "" && user.MobilePhone == string(Phone) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": Phone,
			"password": Password,
		})

		tokenString, error := token.SignedString([]byte("secret"))
		if error != nil {
			doJSONWrite(ctx, 401, "")
			LogController.Error("User not Auth" + Phone)

		}

		LogController.Notice("User Auth" + Phone)
		ctx.Response.Header.Set("Content-Type", "application/json")
		t := tokenString + "@" + reverse(user.Name) + "!3423423423423423jhb234g3h2ik53l5g34h5jgv234oy5gv43jh25v34l5v43uo25vy"

		doJSONWrite(ctx, 200, t)

	} else {
		doJSONWrite(ctx, 401, "")

		LogController.Warning("User not Auth: " + Phone)
	}

}

/*
get all turn for user with token auth

url: http://localhost:8020/user/@token/history
method:GET
params:
@token => eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6IjExZjE5Y2M2Njg2Y2QxOWI1NTM2OTI1YmU3ODNlMDRlZjlmZDViZDk1MTVmMzA1YmQ5OTE4NjZmM2YwMGRmNjAiLCJ1c2VybmFtZSI6IjA5MTIxMjEyMzEifQ.eTH8Qo8P_MZ7Ui1WOm4Cl-9VHLiDiq4Z668cobhJdFM
*/
func GetUserHistoryByToken(ctx *fasthttp.RequestCtx) {
	var queues []models.Queue
	var users []models.User
	var banks []models.Bank
	db := db.GetDB()

	//var Bank models.Bank
	ptoken := ctx.UserValue("token").(string)

	res := string(ValidateMiddleware(string(ptoken)))
	if res != "0" {
		db.Where("mobile_phone = ?", res).First(&users)
		fmt.Println(users[0].ID)

		if err := db.Preload("User").Preload("Bank").Where(&models.Queue{UserID: int(users[0].ID)}).Find(&queues, &banks).Error; err != nil {
			ctx.Error("Not Found", fasthttp.StatusNotFound)
			LogController.Error(string(err.Error()))
		} else {
			jsonValue, err := json.Marshal(queues)
			if err != nil {

			}

			LogController.Notice("Users Fetch History ")
			ctx.Write([]byte(jsonValue))
		}

	} else {
		doJSONWrite(ctx, 401, "")

		LogController.Warning("User not Auth: " + ptoken)
	}

}

/*
last queue number

url: http://localhost:8020/queue/@token
method:GET
params:
@token => eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6IjExZjE5Y2M2Njg2Y2QxOWI1NTM2OTI1YmU3ODNlMDRlZjlmZDViZDk1MTVmMzA1YmQ5OTE4NjZmM2YwMGRmNjAiLCJ1c2VybmFtZSI6IjA5MTIxMjEyMzEifQ.eTH8Qo8P_MZ7Ui1WOm4Cl-9VHLiDiq4Z668cobhJdFM
*/
func GetUserByToken(ctx *fasthttp.RequestCtx) {

	var Queue models.Queue
	var Bank models.Bank
	ptoken := ctx.UserValue("token").(string)

	token, _ := jwt.Parse(string(ptoken), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte("secret"), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user models.User
		mapstructure.Decode(claims, &user)

		id := user.ID
		w64, err := strconv.ParseUint(string(id), 10, 32)
		if err != nil {

		}
		db := db.GetDB()

		if err := db.Preload("User").Preload("Bank").Where(&models.Queue{UserID: int(w64)}).Find(&Queue, &Bank).Error; err != nil {
			ctx.Error("Not Found", fasthttp.StatusNotFound)
			LogController.Error(string(err.Error()))
		} else {
			jsonValue, err := json.Marshal(Queue)
			if err != nil {

			}

			LogController.Notice("User Fetch History ")
			ctx.Write([]byte(jsonValue))
		}

	} else {
		ctx.Write([]byte("401"))
	}

}
