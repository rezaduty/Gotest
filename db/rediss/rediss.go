package rediss

import (
	"fmt"

	"github.com/go-redis/redis"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var rediss *redis.Client
var err error

/*

redis db for store bank viewer

*/

func GetRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	return client
	// Output: PONG <nil>
}
