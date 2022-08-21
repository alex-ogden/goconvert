// GoConvert - a web-based file converter
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
)

type DownloadData struct {
	ImageName string
}

// Set max upload size to be 10MB
const MAX_UPLOAD_SIZE = 10 * 1024 * 1024

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

func runServer(srv_host string) {
	// Define endpoints
	// FileServer
	fileServer := http.FileServer(http.Dir("./static"))
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
	targetDirectory := "static/images"
	// Handle our form upload
	// Max size is ~10MB

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		log.Printf("Upload request was too large!", err)
		http.Error(w, "Upload request was too large!", http.StatusBadRequest)
		return
	}

	// Get file and required format
	format := r.FormValue("imageFormat")
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		log.Printf("Error retrieving file\n")
		log.Print(err, "\n")
		return
	}

	defer file.Close()
	log.Printf("Uploaded File Name: %+v\n", handler.Filename)
	log.Printf("Uploaded File Size: %+vkB\n", (handler.Size / 1000))
	log.Printf("Uploaded File MIME Header: %+v\n", handler.Header)
	log.Printf("Required Image Format: %s\n", format)

	// Create temp file
	targetFileName := fmt.Sprintf("image-%s.%s", fmt.Sprint(rand.Int()), format)
	targetFilePath := fmt.Sprintf("%s/%s", targetDirectory, targetFileName)
	// Read content of uploaded file into byte array
	log.Printf("Reading the contents of the uploaded file into memory\n")
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Print(err, "\n")
	}

	// Convert the image data into other format
	err = convertImage(fileBytes, format, targetFilePath)
	if err != nil {
		log.Print(err, "\n")
	}

	log.Printf("Successfully converted %+v to %s\n", handler.Filename, format)
	log.Printf("Redirecting to download page\n")
	http.Redirect(w, r, fmt.Sprintf("/download?filename=%s", targetFileName), 301)
}

// Converts the image format and returns a slice of bytes
func convertImage(imageBytes []byte, imageFormat, outFilePath string) error {
	// Check the format of the incoming image
	log.Printf("Detecting content type of incoming image\n")
	contentType := http.DetectContentType(imageBytes)

	// Create our outfile
	targetFile, err := os.Create(outFilePath)
	log.Printf("Creating a temporary file at: %s\n", outFilePath)
	if err != nil {
		log.Print(err, "\n")
	}
	defer targetFile.Close()

	// Decide what to do to convert to each format
	switch contentType {
	case "image/png":
		// We have a PNG file
		img, err := png.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			return err
		}

		if imageFormat == "png" {
			log.Printf("Converting %#v to %s\n", contentType, imageFormat)
			if err := png.Encode(targetFile, img); err != nil {
				return err
			}

			return nil
		} else /*More file formats to come*/ {
			log.Printf("Converting %#v to %s\n", contentType, imageFormat)
			if err := jpeg.Encode(targetFile, img, nil); err != nil {
				return err
			}
		}
		// We can have jpegs come in as jpg or jpeg
	case "image/jpg", "image/jpeg":
		// We have a JPG/JPEG file
		img, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			return err
		}

		if imageFormat == "png" {
			log.Printf("Converting %#v to %s\n", contentType, imageFormat)
			if err := png.Encode(targetFile, img); err != nil {
				return err
			}

			return nil
		} else /*More file formats to come*/ {
			log.Printf("Converting %#v to %s\n", contentType, imageFormat)
			if err := jpeg.Encode(targetFile, img, nil); err != nil {
				return err
			}

			return nil
		}
	}
	return fmt.Errorf("Unkown content type. Unable to convert %#v to %s", contentType, imageFormat)
}

// Download the image to the user's computer
func handleDownload(w http.ResponseWriter, r *http.Request) {
	filenameParam := r.URL.Query()["filename"]

	fileName := filenameParam[0]

	log.Printf("Got request to download file: %s", fileName)

	// Finally parse the template for downloading the image back to the user's PC
	tmpl := template.Must(template.ParseFiles("static/download.html"))
	data := DownloadData{
		ImageName: fileName,
	}
	log.Printf("Rendering download page\n")
	tmpl.Execute(w, data)
}

func handleCleanup(w http.ResponseWriter, r *http.Request) {
	log.Print("Got request for cleanup")
	filepathParam := r.URL.Query()["filepath"]
	filePath := filepathParam[0]

	log.Printf("Removing file: %s", filePath)

	if err := os.Remove(filePath); err != nil {
		log.Print(err, "\n")
	}
	http.Redirect(w, r, "/", 301)
}