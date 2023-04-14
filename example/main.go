package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lab5e/gotileserver"
	osmsampletiles "github.com/lab5e/osm-sample-tiles"
)

func main() {
	mux := http.NewServeMux()

	// The host override can be set to an empty string since everything is handled by the server
	if err := gotileserver.RegisterHandler(mux, "", osmsampletiles.SampleOSMTiles()); err != nil {
		log.Fatalf("error registering handler: %v", err)
	}

	fmt.Println("open http://localhost:8080/map/index.html for demo page")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("error serving: %v", err)
	}
}
