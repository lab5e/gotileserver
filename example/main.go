package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/lab5e/gotileserver"
)

const (
	defaultListenAddr = "0.0.0.0:8080"
	readTimeout       = 10 * time.Second
	readHeaderTimeout = 10 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 20 * time.Second
)

var (
	listenAddr string
)

func main() {
	flag.StringVar(&listenAddr, "listen", defaultListenAddr, "HTTP listen addr <host>:<port>")
	flag.Parse()

	mux := http.NewServeMux()

	if err := gotileserver.RegisterHandler(mux, "http://localhost:8080"); err != nil {
		log.Fatalf("error registering handler: %v", err)
	}

	// Handler chain
	handler := handlers.ProxyHeaders(
		handlers.CombinedLoggingHandler(os.Stdout, mux))

	server := http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	fmt.Println("open http://localhost:8080/map/index.html for demo page")
	if err := server.ListenAndServe(); err != nil {
		log.Printf("error serving: %v", err)
	}
}
