package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	imagesDir  = STATIC_DIR + "/images"
	pdfDir     = STATIC_DIR + "/pdf"
	uploadsDir = STATIC_DIR + "/uploads"
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
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
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
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	// Create our required directories
	requiredDirs := [3]string{imagesDir, pdfDir, uploadsDir}

	for _, directory := range requiredDirs {
		createDirectories(directory)
	}

	// Process the request
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		log.Print("Upload request was too large!", err)
		http.Error(w, "Upload request was too large!", http.StatusBadRequest)
		return
	}

	// Get file from request
	file, handler, err := r.FormFile("myFile")
	// Read file into memory
	fileBytes, err := io.ReadAll(file)

	// Work out content type (i.e application/pdf) then form regex to
	// get just the "pdf" part
	contentType := http.DetectContentType(fileBytes)
	targetRegex := regexp.MustCompile("(application|image)/")

	currentFormat := targetRegex.ReplaceAllString(contentType, "")
	requiredFormat := r.FormValue("fileFormat")
	if err != nil {
		log.Printf("Error retrieving file\n")
		log.Fatal(err)
		return
	}

	defer file.Close()

	// Log info about incoming file
	log.Printf("Uploaded File Name: %+v\n", handler.Filename)
	log.Printf("Uploaded File Size: %+vkB\n", (handler.Size / 1000))
	log.Printf("Uploaded File MIME Header: %+v\n", handler.Header)
	log.Printf("Required Format: %s\n", requiredFormat)

	var convertedFilePath string

	switch currentFormat {
	case "pdf":
		// PDF file uploaded
		switch requiredFormat {
		case "pdf":
			convertedFileName := fmt.Sprintf("pdf-%s.%s", string(rand.Int()), requiredFormat)
			convertedFilePath := fmt.Sprintf("%s/%s", pdfDir, convertedFileName)
			err := os.WriteFile(convertedFilePath, fileBytes, 0600)
			if err != nil {
				log.Fatal(err)
			}
		case "png", "jpg", "jpeg":
			// PDF -> Image
			convertedFilePath, err = convertPDFToImage(currentFormat, requiredFormat, uploadsDir, imagesDir, fileBytes)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "png", "jpg", "jpeg":
		// Image file uploaded
		switch requiredFormat {
		case "pdf":
			// Image -> PDF
			convertedFilePath, err = convertImageToPDF(currentFormat, requiredFormat, uploadsDir, pdfDir, fileBytes)
			if err != nil {
				log.Fatal(err)
			}
		case "png", "jpg", "jpeg":
			// Image -> Image
			convertedFilePath, err = convertImage(currentFormat, requiredFormat, imagesDir, fileBytes)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Printf("Successfully converted %+v to %s\n", handler.Filename, requiredFormat)
	log.Printf("Redirecting to download page\n")
	http.Redirect(w, r, fmt.Sprintf("/download?filepath=%s", convertedFilePath), 301)
}

// Download the image to the user's computer
func handleDownload(w http.ResponseWriter, r *http.Request) {
	filepathParam := r.URL.Query()["filepath"]
	filePath := filepathParam[0]
	log.Printf("Recieved filepath: %s", filePath)
	// Get info about filePath var for later on...
	fileInfo, err := os.Stat(filePath)

	// Defining this for later on
	var fileName string

	if fileInfo.IsDir() {
		log.Println("Directory of images, zipping...")
		targetZip := fmt.Sprintf("%s/images.zip", filePath)
		err := zipFiles(filePath, targetZip)
		if err != nil {
			log.Fatal(err)
		}
		fileName := "images.zip"
		// Since we recieved a directory name, we should add the name of the zip
		// to the end of the filePath var so the HTML link can find it
		filePath = filePath + "/images.zip"
		filePath = strings.ReplaceAll(filePath, "../static/", "")
		log.Printf("File to be downloaded: %s", fileName)
	} else {
		// Remove images/ from fileName to get just the name of the image
		// Remove ../static/ from filePath to get directory relative to html file
		filePath = strings.ReplaceAll(filePath, "../static/", "")
		fileName := strings.ReplaceAll(filePath, "images/", "")
		log.Printf("File to be downloaded: %s", fileName)
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got request to download file: %s", filePath)

	// Finally parse the template for downloading the image back to the user's PC
	tmpl := template.Must(template.ParseFiles(STATIC_DIR + "/download.html"))
	data := DownloadData{
		FilePath: filePath,
		FileName: fileName,
	}
	log.Printf("Rendering download page\n")
	if err := tmpl.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}

func handleCleanup(w http.ResponseWriter, r *http.Request) {
	log.Println("Got request for cleanup")

	// Directories to clean - uploads, images and PDF directories
	var dirsToClean = [3]string{STATIC_DIR + "/images", STATIC_DIR + "/pdf", STATIC_DIR + "/uploads"}

	// Find all files in each dir and remove
	for _, directory := range dirsToClean {
		log.Printf("Removing files in directory: %s\n", directory)

		filesToRemove, err := filepath.Glob(directory + "/*")
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range filesToRemove {
			log.Printf("\tRemoving file: %s\n", file)
			if err := os.Remove(file); err != nil {
				log.Fatal(err)
			}
		}
	}
	http.Redirect(w, r, "/", 301)
}
