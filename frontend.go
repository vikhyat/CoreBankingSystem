package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
	"strconv"
)

var redisConnectionShards redis.Conn
var redisErr error

func redisConnection(account int) redis.Conn {
	return redisConnectionShards
}

func depositHandler(w http.ResponseWriter, r *http.Request) {
	account, err := strconv.Atoi(r.FormValue("account"))
	if err != nil {
		log.Fatal(err)
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Fatal(err)
	}

	accountKey := fmt.Sprintf("account:%d", account)

	redisConnection(account).Do("SETNX", accountKey, 0)
	redisConnection(account).Do("INCRBY", accountKey, amount)
	reply, err := redis.Int(redisConnection(account).Do("GET", accountKey))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, "{account: %d, balance: %d}", account, reply)
}

func withdrawHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")
}

func main() {
	redisConnectionShards, redisErr = redis.Dial("tcp", ":6379")
	if redisErr != nil {
		log.Fatal(redisErr)
	}

	http.HandleFunc("/deposit", depositHandler)
	http.HandleFunc("/withdraw", withdrawHandler)

	log.Fatal(http.ListenAndServe(":8341", nil))
}
