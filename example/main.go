package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lab5e/gotileserver"
)

func main() {
	mux := http.NewServeMux()

	if err := gotileserver.RegisterHandler(mux, "http://localhost:8080"); err != nil {
		log.Fatalf("error registering handler: %v", err)
	}

	fmt.Println("open http://localhost:8080/map/index.html for demo page")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("error serving: %v", err)
	}
}
