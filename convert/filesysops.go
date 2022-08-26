package main

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func createDirectories(directory string) {
	log.Printf("Checking for existence of directory: %s\n", directory)
	if _, err := os.Stat(directory); errors.Is(err, os.ErrNotExist) {
		log.Printf("Creating directory %s as it doesn't exist\n", directory)
		err := os.Mkdir(directory, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func zipFiles(source, target string) error {
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set compression
		header.Method = zip.Deflate

		// Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.Contains(info.Name(), ".zip") {
			return nil
		}

		log.Printf("Zipping file: %s", header.Name)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}
