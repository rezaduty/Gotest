package banks

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"

	LogController "../../controllers/logg"
	UserController "../../controllers/users"
	"github.com/valyala/fasthttp"

	"../../models"

	"../../db"
	"../../db/rediss"
)

/*
use in doJSONWrite func
*/
var (
	strContentType     = []byte("Content-Type")
	strApplicationJSON = []byte("application/json")
)

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
cancel turn

url: http://localhost:8020/bank/@bank_id/queue/@turn_number/cancel
method:GET
params:
	@bank_id = bank id
	@turn_number = user turn number in bank queue
	token= eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6ImU3Y2YzZWY0ZjE3YzM5OTlhOTRmMmM2ZjYxMmU4YTg4OGU1YjEwMjY4NzhlNGUxOTM5OGIyM2JkMzhlYzIyMWEiLCJ1c2VybmFtZSI6IjA5MDMzNjA2NzM0In0.ynvqlkGkEiA0RKAlXoiF5fWC_4urZFh_uvej9nYs_dA


*/
func Index(ctx *fasthttp.RequestCtx) {
	doJSONWrite(ctx, 200, "Banki Web Service")

}
func CancelQueue(ctx *fasthttp.RequestCtx) {
	db := db.GetDB()
	var users []models.User
	var queues []models.Queue
	var banks []models.Bank

	//s := strings.Split(ctx.URI().String(), "/")

	bid := ctx.UserValue("id")
	qid := ctx.UserValue("qid")

	ptoken := ctx.FormValue("token")
	res := string(UserController.ValidateMiddleware(string(ptoken)))
	fmt.Println(res)
	if res != "0" {

		db.Where("mobile_phone = ?", res).First(&users)
		cc := users[0].CC
		cc++
		db.Where("id = ?", bid).First(&banks)
		active_turn_number := banks[0].Active_turns_number
		active_turn_number--
		uid := users[0].ID
		db.Model(users).Where("id = ?", uid).Update("cc", cc)
		db.Model(banks).Where("id = ?", bid).Update("active_turns_number", active_turn_number)
		db.Model(queues).Where("id = ?", qid).Update("cancel", true)
		doJSONWrite(ctx, 200, "")
		LogController.Error("User cancel turn in queue: " + users[0].Email)
	} else {
		doJSONWrite(ctx, 401, "")
		LogController.Error("User not auth in give turn: " + users[0].Email)
	}
}

/*

for bank operator run queue

url: http://localhost:8020/banks/@bank_id/queue/run
method:GET
params:
	@bank_id = bank id
	token= eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6ImU3Y2YzZWY0ZjE3YzM5OTlhOTRmMmM2ZjYxMmU4YTg4OGU1YjEwMjY4NzhlNGUxOTM5OGIyM2JkMzhlYzIyMWEiLCJ1c2VybmFtZSI6IjA5MDMzNjA2NzM0In0.ynvqlkGkEiA0RKAlXoiF5fWC_4urZFh_uvej9nYs_dA



*/
func RunQueue(ctx *fasthttp.RequestCtx) {
	//TODO banks operator token
	db := db.GetDB()
	var banks []models.Bank
	var queues []models.Queue
	var users []models.User

	ptoken := ctx.FormValue("token")
	res := string(UserController.ValidateMiddleware(string(ptoken)))
	fmt.Println(res)
	if res != "0" {
		db.Where("mobile_phone = ?", res).First(&users)
		if users[0].Role == "BO" {
			bid := ctx.UserValue("id")
			db.Where("id = ?", bid).First(&banks)

			active_turns_number := int(banks[0].Active_turns_number)
			active_turns_number--

			db.Where("turn_number = ?", active_turns_number).First(&queues)
			if len(queues) != 0 {
				turn_number := int(queues[0].Turn_number)
				cc := queues[0].Cancel

				if active_turns_number == turn_number {
					if cc {
						active_turns_number--
						db.Model(banks).Where("id = ?", bid).Update("active_turns_number", active_turns_number)
						doJSONWrite(ctx, 200, "User with email: "+users[0].Email+" should now in bank with id: "+string(banks[0].ID))
						LogController.Notice("User with email: " + users[0].Email + " should now in bank with id: " + string(banks[0].ID))
					} else {
						doJSONWrite(ctx, 200, "User with email: "+users[0].Email+" should now in bank with id: "+string(banks[0].ID))
						LogController.Notice("User with email: " + users[0].Email + " should now in bank with id: " + string(banks[0].ID))
						db.Model(banks).Where("id = ?", bid).Update("active_turns_number", active_turns_number)
					}
				} else {
					doJSONWrite(ctx, 200, "No User with this turn_number: "+string(turn_number))
					LogController.Warning("No User with this turn_number: " + string(turn_number))
					db.Model(banks).Where("id = ?", bid).Update("active_turns_number", active_turns_number)
				}
			} else {
				doJSONWrite(ctx, 404, "Not any person in queue")
				LogController.Warning("Not any person in queue: " + string(banks[0].ID))
			}

		} else {
			doJSONWrite(ctx, 401, "")
			LogController.Error("Your Dont Permissin Run Queue: " + string(users[0].Email))
		}

	} else {
		doJSONWrite(ctx, 401, "")
		LogController.Error("User not auth in give turn: " + string(ptoken))
	}

}

/*
for add user to queue

url: http://localhost:8020/banks/@id/queue/add
method:GET
params:
	@id= bank id
	token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6IjExZjE5Y2M2Njg2Y2QxOWI1NTM2OTI1YmU3ODNlMDRlZjlmZDViZDk1MTVmMzA1YmQ5OTE4NjZmM2YwMGRmNjAiLCJ1c2VybmFtZSI6IjA5MTIxMjEyMzEifQ.eTH8Qo8P_MZ7Ui1WOm4Cl-9VHLiDiq4Z668cobhJdFM


*/
func AddQueue(ctx *fasthttp.RequestCtx) {
	db := db.GetDB()
	var banks []models.Bank
	var users []models.User

	ptoken := ctx.FormValue("token")
	res := string(UserController.ValidateMiddleware(string(ptoken)))
	fmt.Println(res)
	if res != "0" {
		db.Where("mobile_phone = ?", res).First(&users)
		cc := users[0].CC
		if cc <= 3 {
			user_status := users[0].ActiveStatus
			if user_status <= 20 {
				user_status++

				db.Where("mobile_phone = ?", res).First(&users)

				user_id := users[0].ID

				bid := ctx.UserValue("id").(string)
				db.Where("id = ?", bid).First(&banks)
				active_turns_number := int(banks[0].Active_turns_number)
				fmt.Println(active_turns_number)
				turns_number := banks[0].Turns_number
				turns_number++
				db.Model(banks).Where("id = ?", bid).Update(models.Bank{Turns_number: turns_number, Active_turns_number: turns_number})

				bank_id, err := strconv.Atoi(bid)
				if err != nil {
					LogController.Error(err.Error())
				}

				queues := &models.Queue{BankID: bank_id, UserID: int(user_id), Turn_number: turns_number, Turn_status: 2, Method: 1, Temp_Active_Turn_Number: turns_number}

				db.Create(&queues)

				jsonValue, err := json.Marshal(queues)
				if err != nil {
					LogController.Error(err.Error())
				}

				db.Model(users).Where("id = ?", int(user_id)).Update("active_status", user_status)
				ctx.Response.Header.Set("Content-Type", "application/json")
				ctx.Write([]byte(jsonValue))

			} else {
				doJSONWrite(ctx, 208, "User has max give queue per day")
				LogController.Warning("User has max give queue per day: " + users[0].Email)
			}
		} else {
			doJSONWrite(ctx, 208, "User has max cancel queue per day")
			LogController.Warning("User has max cancel queue per day: " + users[0].Email)
		}

	} else {
		doJSONWrite(ctx, 401, "")
		LogController.Error("User not auth in give turn: " + users[0].Email)
	}

}

/*

lng and lat for sort bank by near bank
if lng and lat is null then sort by id desc

url: http://localhost:8020/banks
method:GET
params:
	lng=51.32748270467971
	lat=35.74402097029468

*/

func GetBanks(ctx *fasthttp.RequestCtx) {

	lng := ctx.FormValue("lng")
	lat := ctx.FormValue("lat")
	if len(string(lng)) < 2 || len(string(lat)) < 2 {
		lng = []byte("1")
		lat = []byte("1")
	}
	var banks []models.Bank
	db := db.GetDB()

	if err := db.Order("((x-" + string(lat) + ")*(x-" + string(lat) + ")) + ((y - " + string(lng) + ")*(y - " + string(lng) + ")) ASC").Find(&banks).Error; err != nil {

		fmt.Println(err)
	} else {
		jsonValue, err := json.Marshal(banks)
		if err != nil {

		}
		LogController.Notice("Banks Fetch")
		ctx.Write([]byte(jsonValue))
	}

}

/*

get bank with id

*/
func GetBankById(ctx *fasthttp.RequestCtx) {

	var banks []models.Bank
	id := ctx.UserValue("id")

	db := db.GetDB()
	client := rediss.GetRedis()
	val, err := client.Get(id.(string)).Result()
	if err != nil {
		panic(err)
	}
	LogController.Notice("Bank " + id.(string) + "Fetch viewer " + val)

	t, er := strconv.Atoi(val)
	if er != nil {
		// handle error
	}
	t = t + 1
	err = client.Set(id.(string), strconv.Itoa(t), 0).Err()
	if err != nil {
		panic(err)
	}

	if err := db.Where("Id = ?", id.(string)).First(&banks).Error; err != nil {

		fmt.Println(err)
	} else {
		jsonValue, err := json.Marshal(banks)
		if err != nil {

		}
		var jsonStr = []byte(`{"data":` + string(jsonValue) + `,"viewer":` + strconv.Itoa(t) + `}`)
		ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
		ctx.Response.SetStatusCode(200)
		ctx.Write([]byte(jsonStr))

		LogController.Notice("Bank Fetch with " + id.(string))

	}

}
