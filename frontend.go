package main

import (
	"fmt"
	"log"
	"net/http"
)

func depositHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")
}

func main() {
	http.HandleFunc("/deposit", depositHandler)

	log.Fatal(http.ListenAndServe(":8341", nil))
}
