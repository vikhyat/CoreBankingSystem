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

func accountKey(account int) string {
	return fmt.Sprintf("account:%d", account)
}

func getBalance(account int) int {
	redisConnection(account).Do("SETNX", accountKey(account), 0)
	balance, err := redis.Int(redisConnection(account).Do("GET", accountKey(account)))
	if err != nil {
		log.Fatal(err)
	}
	return balance
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

	redisConnection(account).Do("SETNX", accountKey(account), 0)
	redisConnection(account).Do("INCRBY", accountKey(account), amount)

	fmt.Fprintf(w, "{account: %d, balance: %d}", account, getBalance(account))
}

func withdrawHandler(w http.ResponseWriter, r *http.Request) {
	account, err := strconv.Atoi(r.FormValue("account"))
	if err != nil {
		log.Fatal(err)
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Fatal(err)
	}

	if getBalance(account) >= amount {
		redisConnection(account).Do("DECRBY", accountKey(account), amount)
		fmt.Fprintf(w, "{account: %d, balance: %d}", account, getBalance(account))
	} else {
		fmt.Fprintf(w, "{error: \"insufficient funds\"}")
	}
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
