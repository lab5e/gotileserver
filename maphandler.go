package main

import (
	"embed" // For embedding the index page
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	_ "modernc.org/sqlite" // use sqlite driver
)

//go:embed map/index.html
var indexPage []byte

//go:embed map/tiles
var tiles embed.FS

//go:embed map/fonts
var fonts embed.FS

//go:embed map/styles
var styles embed.FS

func main() {
	mux := http.NewServeMux()
	RegisterHandler(mux, "http://localhost")
	fmt.Println("Serving...")
	if err := http.ListenAndServe("0.0.0.0:80", handlers.LoggingHandler(os.Stdout, mux)); err != nil {
		fmt.Println("Error serving: ", err.Error())
	}
}

// RegisterHandler registers handlers for the /map path in the mux. Since the styles and spec files require
// a fair bit of massaging to work. Omit the trailing slash for the hostname (ie. "http://localhost")
func RegisterHandler(mux *http.ServeMux, hostName string) {

	mux.HandleFunc("/map/index.html", static("text/html", indexPage))
	mux.HandleFunc("/map/tiles/", tileHandler)
	mux.Handle("/map/fonts/", http.FileServer(http.FS(fonts)))
	mux.Handle("/map/styles/", http.FileServer(http.FS(styles)))

}

func static(c string, v []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", c)
		w.WriteHeader(http.StatusOK)
		w.Write(v)
	}
}

func tileHandler(w http.ResponseWriter, r *http.Request) {
	var data []byte
	var err error
	fileName := r.URL.Path[1:]
	data, err = tiles.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error retrieving tile %s. Serving empty tile\n", fileName)
		data, err = hex.DecodeString("1F8B0800FA78185E000393E2E3628F8F4FCD2D28A9D46850A86002006471443610000000")
		if err != nil {
			fmt.Printf("Empty tile error: %v\n", err)
		}
	}
	if strings.HasSuffix(fileName, ".json") {
		// This is the metadata; report appropriate content type
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	w.Header().Add("Content-Type", "application/x-protobuf")
	w.Header().Add("Content-Encoding", "gzip") // The generated tiles are gzipped by default
	w.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
