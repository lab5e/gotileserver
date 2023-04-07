package main

import (
	"fmt"
	"net/http"

	"github.com/lab5e/gotileserver"
)

func main() {
	mux := http.NewServeMux()
	if err := gotileserver.RegisterHandler(mux, "http://localhost:8080"); err != nil {
		fmt.Printf("Error registering handler: %v\n", err)
		return
	}
	fmt.Println("Serving... open http://localhost:8080/map/index.html for demo page")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		fmt.Println("Error serving: ", err.Error())
	}
}
