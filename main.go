package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/matiasinsaurralde/product-services/api"
)

const (
	defaultListenAddr = ":9999"
)

func main() {
	log.Println("Initializing payments service")
	// By default grab the current working directory
	// and use the "data" subdirectory as payment service base path:
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dataPath := filepath.Join(cwd, "data")
	log.Printf("Setting data directory to '%s'\n", dataPath)

	// Initialize the API and start the HTTP server:
	apiHandler, err := api.NewHandler(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := http.ListenAndServe(defaultListenAddr, apiHandler); err != nil {
		log.Fatal(err)
	}
}
