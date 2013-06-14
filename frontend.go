package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
)

var redisConnection redis.Conn
var redisErr error

func depositHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")
}

func main() {
	redisConnection, redisErr = redis.Dial("tcp", ":6379")
	if redisErr != nil {
		log.Fatal(redisErr)
	}

	http.HandleFunc("/deposit", depositHandler)
	log.Fatal(http.ListenAndServe(":8341", nil))
}
