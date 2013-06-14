package main

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
	"strconv"
	"time"
)

const redisInstancesCount = 1

var redisConnectionShards [redisInstancesCount]redis.Conn
var redisErr error

// Return a connection to the Redis instance that is responsible for this account.
func redisConnection(account int) redis.Conn {
	return redisConnectionShards[account%redisInstancesCount]
}

// Return the key that will store the account balance.
func accountKey(account int) string {
	return fmt.Sprintf("account:%d", account)
}

// Return the lock key for the given account.
func accountLockKey(account int) string {
	return fmt.Sprintf("accountlock:%d", account)
}

// Return the balance of an account.
func getBalance(account int) int {
	redisConnection(account).Do("HSETNX", "accounts", account, 0)
	balance, err := redis.Int(redisConnection(account).Do("HGET", "accounts", account))
	if err != nil {
		log.Print(err)
	}
	return balance
}

// Acquire the lock on the given account.
func acquireLock(account int) error {
	retries := 0

	for retries <= 100 {
		reply, _ := redis.Int(redisConnection(account).Do("SETNX", accountLockKey(account), "LOCKED"))

		if reply == 1 {
			redisConnection(account).Do("EXPIRE", accountLockKey(account), 5)
			return nil
		} else {
			time.Sleep(100 * time.Millisecond)
			retries += 1
			if retries == 10 {
				return errors.New("could not acquire lock")
			}
		}
	}

	return nil
}

// Release the lock on the given account.
func releaseLock(account int) {
	redisConnection(account).Do("DEL", accountLockKey(account))
}

func depositHandler(w http.ResponseWriter, r *http.Request) {
	account, err := strconv.Atoi(r.FormValue("account"))
	if err != nil {
		log.Print(err)
	}

	if err := acquireLock(account); err != nil {
		fmt.Fprintf(w, "{error: \"could not acquire lock\"}")
		return
	}
	defer releaseLock(account)

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Print(err)
	}

	getBalance(account)
	redisConnection(account).Do("HINCRBY", "accounts", account, amount)

	fmt.Fprintf(w, "{account: %d, balance: %d}", account, getBalance(account))
}

func withdrawHandler(w http.ResponseWriter, r *http.Request) {
	account, err := strconv.Atoi(r.FormValue("account"))
	if err != nil {
		log.Print(err)
	}

	if err := acquireLock(account); err != nil {
		fmt.Fprintf(w, "{error: \"could not acquire lock\"}")
		return
	}
	defer releaseLock(account)

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Print(err)
	}

	if getBalance(account) >= amount {
		redisConnection(account).Do("HINCRBY", "accounts", account, -amount)
		fmt.Fprintf(w, "{account: %d, balance: %d}", account, getBalance(account))
	} else {
		fmt.Fprintf(w, "{error: \"insufficient funds\"}")
	}
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	source, err := strconv.Atoi(r.FormValue("source"))
	if err != nil {
		log.Print(err)
	}

	if err := acquireLock(source); err != nil {
		fmt.Fprintf(w, "{error: \"could not acquire lock\"}")
		return
	}
	defer releaseLock(source)

	destination, err := strconv.Atoi(r.FormValue("destination"))
	if err != nil {
		log.Print(err)
	}

	if err := acquireLock(destination); err != nil {
		fmt.Fprintf(w, "{error: \"could not acquire lock\"}")
		return
	}
	defer releaseLock(destination)

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Print(err)
	}

	if getBalance(source) >= amount {
		getBalance(destination)
		redisConnection(source).Do("HDECRBY", "accounts", source, amount)
		redisConnection(destination).Do("HINCRBY", "accounts", destination, amount)
		fmt.Fprintf(w, "{success: \"ok\"}")
	} else {
		fmt.Fprintf(w, "{error: \"insufficient funds\"}")
	}
}

func main() {
	redisInstances := [redisInstancesCount]string{"127.0.0.1:6379"}

	for i, s := range redisInstances {
		redisConnectionShards[i], redisErr = redis.Dial("tcp", s)
		if redisErr != nil {
			log.Print(redisErr)
		}
	}

	http.HandleFunc("/deposit", depositHandler)
	http.HandleFunc("/withdraw", withdrawHandler)
	http.HandleFunc("/transfer", transferHandler)

	log.Print(http.ListenAndServe(":8341", nil))
}
