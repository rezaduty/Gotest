package main

import (
	"log"

	LogController "./controllers/logg"

	BankController "./controllers/banks"
	UserController "./controllers/users"

	"./db"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func main() {
	//TODO user profile
	url := "localhost"
	port := "8020"
	LogController.Info("Serve in " + url + ":" + port)

	db.Init()

	router := fasthttprouter.New()
	router.GET("/", BankController.Index)

	router.GET("/queue/:token", UserController.GetUserByToken)
	router.GET("/user/:token/history", UserController.GetUserHistoryByToken)

	router.POST("/user/login", UserController.Login)
	router.POST("/user/register", UserController.Register)

	router.GET("/banks", BankController.GetBanks)
	router.GET("/banks/:id", BankController.GetBankById)
	router.GET("/banks/:id/queue/add", BankController.AddQueue)
	router.GET("/banks/:id/queue/run", BankController.RunQueue)
	router.GET("/bank/:id/queue/:qid/cancel", BankController.CancelQueue)
	log.Fatal(fasthttp.ListenAndServe(":"+port, router.Handler))

}
