// GoConvert - a web-based file converter
package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// Converts the image format and returns a slice of bytes
func convertImage(imageBytes []byte, imageFormat, outFilePath string) error {
	// Check the format of the incoming image
	log.Printf("Detecting content type of incoming image\n")
	contentType := http.DetectContentType(imageBytes)

	// Create our outfile
	targetFile, err := os.Create(outFilePath)
	log.Printf("Creating a temporary file at: %s\n", outFilePath)
	if err != nil {
		return err
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

func convertPDFToImage(filename, desiredFormat string, uploadedFile multipart.File) (string, error) {
	// Take the uploaded file, read it into memory and write that to disk
	pdfDirectory := STATIC_DIR + "/pdf"
	pdfFileName := "upload.pdf"
	pdfFullPath := pdfDirectory + "/" + pdfFileName
	if _, err := os.Stat(pdfDirectory); errors.Is(err, os.ErrNotExist) {
		log.Printf("Creating directory %s as it doesn't exist\n", pdfDirectory)
		err := os.Mkdir(pdfDirectory, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	// Set values for the image file to follow
	imageDirectory := STATIC_DIR + "/images"
	imageName := fmt.Sprintf("upload-%s.%s", fmt.Sprint(rand.Int()), desiredFormat)
	imageFullPath := imageDirectory + "/" + imageName
	if _, err := os.Stat(imageDirectory); errors.Is(err, os.ErrNotExist) {
		log.Printf("Creating directory %s as it doesn't exist\n", imageDirectory)
		err := os.Mkdir(imageDirectory, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	defer uploadedFile.Close()

	// Read the uploaded (PDF) file into memory
	log.Println("Reading uploaded PDF into memory")
	fileBytes, err := io.ReadAll(uploadedFile)
	if err != nil {
		return "", err
	}

	// Write the PDF file to disk
	log.Println("Writing the uploaded PDF file to disk")
	if err := os.WriteFile(pdfFullPath, fileBytes, 0600); err != nil {
		return "", err
	}

	cmd := exec.Command("identify", "-format", "%n", pdfFullPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	numPages, err := strconv.ParseInt(string(out.String()), 10, 32)
	if err != nil {
		return "", err
	}

	log.Printf("We have %s pages in the PDF\n", string(out.String()))

	// Initialise imagemagick
	log.Println("Initialising imagemagick")
	imagick.Initialize()
	defer imagick.Terminate()

	// Create a new wand
	log.Println("Creating a new wand")
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	log.Println("Setting high resolution for resulting image")
	if err := mw.SetResolution(100, 100); err != nil {
		return "", err
	}

	// Read the PDF file into memory (we have to do this as imagemagick won't take a slice of bytes, it has to be a written file)
	log.Printf("Reading the PDF %s into memory\n", pdfFullPath)
	if err := mw.ReadImage(pdfFullPath); err != nil {
		return "", err
	}

	if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_REMOVE); err != nil {
		return "", err
	}

	// Set any compression (100 = max quality)
	if err := mw.SetCompressionQuality(100); err != nil {
		return "", err
	}

	for i := int64(0); i < numPages; i++ {
		log.Printf("Converting page #%s", fmt.Sprint(i))
		mw.SetIteratorIndex(int(i)) // This being the page offset

		log.Printf("Setting the wand to write file as format: %s", desiredFormat)
		if err := mw.SetImageFormat(desiredFormat); err != nil {
			return "", err
		}

		log.Printf("Writing file: %s", imageFullPath)
		if err := mw.WriteImage(imageFullPath); err != nil {
			return "", err
		}
	}
	return imageName, nil
}
