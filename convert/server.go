package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func runServer(srv_host string) {
	// Define endpoints
	// FileServer
	fileServer := http.FileServer(http.Dir(STATIC_DIR))
	http.Handle("/", fileServer)
	http.HandleFunc("/health", handleHealthCheck)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/download", handleDownload)
	http.HandleFunc("/cleanup", handleCleanup)
	// Start the server
	log.Printf("Starting web server on: %s\n", srv_host)
	if err := http.ListenAndServe(srv_host, nil); err != nil {
		log.Fatal("Could not start server: ", err)
	}
}

// Health check
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("Request method recieved was not GET\n")
		log.Printf("Request method: %s\n", r.Method)
		http.Error(w, "Method not supported", http.StatusNotFound)
		return
	}

	log.Printf("Got request for /health - responding with Healthy")
	fmt.Fprintf(w, "Healthy")
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Make sure we only handle POST requests
	if r.Method != "POST" {
		log.Printf("Request method recieved was not POST\n")
		log.Printf("Request method: %s\n", r.Method)
		http.Error(w, "Method not supported", http.StatusNotFound)
		return
	}

	log.Printf("New request for file upload\n")
	targetDirectory := STATIC_DIR + "/images"

	log.Printf("Checking for existence of directory: %s\n", targetDirectory)
	if _, err := os.Stat(targetDirectory); errors.Is(err, os.ErrNotExist) {
		log.Printf("Creating directory %s as it doesn't exist\n", targetDirectory)
		err := os.Mkdir(targetDirectory, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Handle our form upload
	// Max size is ~10MB
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		log.Print("Upload request was too large!", err)
		http.Error(w, "Upload request was too large!", http.StatusBadRequest)
		return
	}

	// Get file and required format
	format := r.FormValue("imageFormat")
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		log.Printf("Error retrieving file\n")
		log.Fatal(err)
		return
	}

	defer file.Close()
	log.Printf("Uploaded File Name: %+v\n", handler.Filename)
	log.Printf("Uploaded File Size: %+vkB\n", (handler.Size / 1000))
	log.Printf("Uploaded File MIME Header: %+v\n", handler.Header)
	log.Printf("Required Format: %s\n", format)

	// Create temp file
	targetFileName := fmt.Sprintf("image-%s.%s", fmt.Sprint(rand.Int()), format)
	targetFilePath := fmt.Sprintf("%s/%s", targetDirectory, targetFileName)

	// Convert the image data into other format
	if strings.Contains(handler.Filename, ".pdf") {
		log.Println("A PDF file has been uploaded")
		// Uploaded file is a PDF file
		targetFileName, err = convertPDFToImage(handler.Filename, format, file)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("An image file has been uploaded")
		// Read content of uploaded file into byte array
		log.Printf("Reading the contents of the uploaded file into memory\n")
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		err = convertImage(fileBytes, format, targetFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Successfully converted %+v to %s\n", handler.Filename, format)
	log.Printf("Redirecting to download page\n")
	http.Redirect(w, r, fmt.Sprintf("/download?filename=%s", targetFileName), 301)
}

// Download the image to the user's computer
func handleDownload(w http.ResponseWriter, r *http.Request) {
	filenameParam := r.URL.Query()["filename"]
	fileName := filenameParam[0]
	log.Printf("Got request to download file: %s", fileName)

	// Finally parse the template for downloading the image back to the user's PC
	tmpl := template.Must(template.ParseFiles(STATIC_DIR + "/download.html"))
	data := DownloadData{
		ImageName: fileName,
	}
	log.Printf("Rendering download page\n")
	if err := tmpl.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}

func handleCleanup(w http.ResponseWriter, r *http.Request) {
	log.Print("Got request for cleanup")

	// Directories to clean - images and PDF directories
	var dirsToClean = [2]string{STATIC_DIR + "/images", STATIC_DIR + "/pdf"}

	// Find all files in each dir and remove
	for _, directory := range dirsToClean {
		log.Printf("Removing files in directory: %s", directory)

		filesToRemove, err := filepath.Glob(directory + "/*")
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range filesToRemove {
			log.Printf("\tRemoving file: %s", file)
			if err := os.Remove(file); err != nil {
				log.Fatal(err)
			}
		}
	}
	http.Redirect(w, r, "/", 301)
}
