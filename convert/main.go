package main

import (
	"log"
	"os"
)

type DownloadData struct {
	FilePath string
	FileName string
}

// Set max upload size to be 50MB
const MAX_UPLOAD_SIZE = 50 * 1024 * 1024
const STATIC_DIR = "../static"

func main() {
	// Define the port for our server to run on
	srv_host := "0.0.0.0"
	srv_port := os.Getenv("PORT")
	// If we don't have the PORT env var, default to port 4433
	if srv_port == "" {
		log.Print("Couldn't find environment variable: PORT\n")
		log.Print("Defaulting to port: 4433\n")
		srv_port = "4433"
	}

	srv_host = srv_host + ":" + srv_port
	runServer(srv_host)
}
