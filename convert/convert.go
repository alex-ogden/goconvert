// GoConvert - a web-based file converter
package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"

	"github.com/br3w0r/goitopdf/itopdf"
	"gopkg.in/gographics/imagick.v3/imagick"
)

// Converts the image format and returns a slice of bytes
func convertImage(currentFormat, requiredFormat, imagesDir string, fileBytes []byte) (string, error) {
	// Create our outfile
	convertedFileName := fmt.Sprintf("image-%s.%s", fmt.Sprint(rand.Int()), requiredFormat)
	convertedFilePath := fmt.Sprintf("%s/%s", imagesDir, convertedFileName)
	convertedFile, err := os.Create(convertedFilePath)
	if err != nil {
		return "", err
	}

	defer convertedFile.Close()

	// Decide what to do to convert to each format
	switch currentFormat {
	case "png":
		// We have a PNG file
		log.Println("Decoding uploaded PNG file")
		decodedFile, err := png.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			return "", err
		}

		log.Printf("Converting %#v to %s\n", currentFormat, requiredFormat)
		if requiredFormat == "png" {
			err := png.Encode(convertedFile, decodedFile)
			if err != nil {
				return "", err
			}

			return convertedFilePath, nil
		} else {
			err := jpeg.Encode(convertedFile, decodedFile, nil)
			if err != nil {
				return "", err
			}

			return convertedFilePath, nil
		}
		// We can have jpegs come in as jpg or jpeg
	case "jpg", "jpeg":
		// We have a JPG/JPEG file
		log.Println("Decoding uploaded JPEG/JPG file")
		decodedFile, err := jpeg.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			return "", err
		}

		log.Printf("Converting %#v to %s\n", currentFormat, requiredFormat)
		if requiredFormat == "png" {
			err := png.Encode(convertedFile, decodedFile)
			if err != nil {
				return "", err
			}

			return convertedFilePath, nil
		} else {
			err := jpeg.Encode(convertedFile, decodedFile, nil)
			if err != nil {
				return "", err
			}

			return convertedFilePath, nil
		}
	}
	return "", fmt.Errorf("Unkown content type. Unable to convert %#v to %s", currentFormat, requiredFormat)
}

func convertPDFToImage(currentFormat, requiredFormat, uploadsDir, imagesDir string, fileBytes []byte) (string, error) {
	// Create our outfile
	currentFileName := fmt.Sprintf("upload-%s.%s", fmt.Sprint(rand.Int()), currentFormat)
	currentFilePath := fmt.Sprintf("%s/%s", uploadsDir, currentFileName)

	// Write the PDF file to disk
	log.Println("Writing the uploaded PDF file to disk")
	err := os.WriteFile(currentFilePath, fileBytes, 0600)
	if err != nil {
		return "", err
	}

	// Work out the number of pages in the PDF document
	log.Println("Working out the number of pages in the PDF file")
	cmd_output, err := exec.Command("convert", currentFilePath, "-set", "option:totpages", "%[n]", "-delete", "1--1", "-format", "%[totpages]", "info:").Output()
	if err != nil {
		return "", err
	}
	log.Printf("We have %s pages in the PDF\n", cmd_output)
	numPages, err := strconv.ParseInt(string(cmd_output), 10, 32)

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
	log.Printf("Reading the PDF %s into memory\n", currentFilePath)
	if err := mw.ReadImage(currentFilePath); err != nil {
		return "", err
	}

	log.Println("Instructing wand to remove alpha channel")
	if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_REMOVE); err != nil {
		return "", err
	}

	// Set any compression (100 = max quality)
	log.Println("Instructing wand to set maximum quality")
	if err := mw.SetCompressionQuality(100); err != nil {
		return "", err
	}

	// Generate a random number for the file(s)
	randomNumber := fmt.Sprint(rand.Int())

	for i := int64(0); i < numPages; i++ {
		convertedFileName := fmt.Sprintf("image-%s-%s.%s", randomNumber, fmt.Sprint(i), requiredFormat)
		convertedFilePath := fmt.Sprintf("%s/%s", imagesDir, convertedFileName)

		log.Printf("Converting page #%s", fmt.Sprint(i))
		mw.SetIteratorIndex(int(i)) // This being the page offset

		log.Printf("Setting the wand to write file as format: %s", requiredFormat)
		if err := mw.SetImageFormat(requiredFormat); err != nil {
			return "", err
		}

		log.Printf("Writing file: %s", convertedFilePath)
		err := mw.WriteImage(convertedFilePath)
		if err != nil {
			return "", err
		}
	}
	return imagesDir, nil
}

func convertImageToPDF(currentFormat, requiredFormat, uploadsDir, pdfDir string, fileBytes []byte) (string, error) {
	convertedFileName := fmt.Sprintf("pdf-%s.%s", fmt.Sprint(rand.Int()), requiredFormat)
	convertedFilePath := fmt.Sprintf("%s/%s", pdfDir, convertedFileName)

	log.Printf("Creating a temporary file at: %s\n", convertedFilePath)
	convertedFile, err := os.Create(convertedFilePath)
	if err != nil {
		return "", err
	}

	defer convertedFile.Close()

	switch currentFormat {
	case "png":
		log.Println("Decoding the PNG data")
		fileDecoded, err := png.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			return "", err
		}

		log.Println("Re-encoding PNG data and writing to disk")
		err = png.Encode(convertedFile, fileDecoded)
		if err != nil {
			return "", err
		}
	case "jpg", "jpeg":
		log.Println("Decoding the JPEG data")
		fileDecoded, err := jpeg.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			return "", err
		}

		log.Println("Re-encoding JPEG data and writing to disk")
		err = jpeg.Encode(convertedFile, fileDecoded, nil)
		if err != nil {
			return "", err
		}
	}

	log.Println("Creating new ImageToPDF instance")
	pdf := itopdf.NewInstance()

	log.Println("Detecting image file to convert to PDF")
	err = pdf.AddImage(convertedFile.Name())
	if err != nil {
		return "", err
	}

	log.Println("Saving image file as PDF")
	err = pdf.Save(convertedFilePath)
	if err != nil {
		return "", err
	}

	return convertedFilePath, nil
}
