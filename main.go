package main

import (
	"database/sql"
	_ "embed" // For embedding the index page
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
	_ "modernc.org/sqlite" // use sqlite driver
)

//go:embed index.html
var indexPage []byte

//go:embed style.json
var styleJson []byte

//go:embed spec.json
var specJson []byte

//go:embed osm_liberty_local.json
var libertyJson []byte

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "trondheim-osm.mbtiles")
	if err != nil {
		fmt.Printf("Error opening tile DB: %v\n", err)
		return
	}
	fmt.Println("Opened tile db")

	mux := http.NewServeMux()
	mux.HandleFunc("/index.html", static("text/html", indexPage))
	mux.HandleFunc("/style.json", static("text/json", styleJson))
	mux.HandleFunc("/spec.json", static("text/json", specJson))
	mux.HandleFunc("/osm_liberty_local.json", static("text/json", libertyJson))
	mux.HandleFunc("/metadata", metadataHandler)
	mux.HandleFunc("/tiles/", tileHandler)
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))

	if err := http.ListenAndServe("0.0.0.0:80", handlers.LoggingHandler(os.Stdout, mux)); err != nil {
		fmt.Println("Error serving: ", err.Error())
	}
}

func static(c string, v []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", c)
		w.WriteHeader(http.StatusOK)
		w.Write(v)
	}
}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT name, value FROM metadata")
	if err != nil {
		fmt.Printf("Error reading metadata: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	ret := make(map[string]interface{})
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			fmt.Printf("error scanning metdata: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if k == "json" {
			buf := strings.NewReader(v)
			content := make(map[string]interface{})
			json.NewDecoder(buf).Decode(&content)
			ret[k] = content
			continue
		}
		ret[k] = v
	}

	json.NewEncoder(w).Encode(ret)
}
func tileHandler(w http.ResponseWriter, r *http.Request) {
	// Slight hackish path extraction but we'll get /tiles/x/y/z.pbf
	elements := strings.Split(r.URL.Path, "/")
	if len(elements) != 5 {
		fmt.Printf("Invalid tile path, return 400")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	z, err := strconv.ParseInt(elements[2], 10, 32)
	if err != nil {
		fmt.Println("Invalid X value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	x, err := strconv.ParseInt(elements[3], 10, 32)
	if err != nil {
		fmt.Println("Invalid Y value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	file := elements[4]
	y, err := strconv.ParseInt(strings.Split(file, ".")[0], 10, 32)
	if err != nil {
		fmt.Println("Invalid Z value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// I haven't seen a good explanation for this.
	y = ((1 << z) - y - 1)
	//	fmt.Printf("X: %d Y: %d Z: %d  File: %s\n", x, y, z, file)

	row := db.QueryRow("SELECT tile_data FROM tiles WHERE zoom_level = $1 AND tile_column = $2 AND tile_row = $3", z, x, y)
	var data []byte
	if err := row.Scan(&data); err != nil {
		fmt.Printf("Not found: %v\n", err)
		data, err = hex.DecodeString("1F8B0800FA78185E000393E2E3628F8F4FCD2D28A9D46850A86002006471443610000000")
		if err != nil {
			fmt.Printf("Empty tile error: %v\n", err)
		}
	}
	w.Header().Add("Content-Type", "application/x-protobuf")
	w.Header().Add("Content-Encoding", "gzip") // The generated tiles are gzipped by default
	w.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
