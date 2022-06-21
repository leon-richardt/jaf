package main

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

type uploadHandler struct {
	config *Config
}

func (h *uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	uploadFile, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "could not read uploaded file: "+err.Error(), http.StatusBadRequest)
		log.Println("    could not read uploaded file: " + err.Error())
		return
	}
	defer uploadFile.Close()

	_, fileExtension := splitFileName(header.Filename)
	link, err := generateLink(h, &uploadFile, fileExtension)
	if err != nil {
		http.Error(w, "could not save file: "+err.Error(), http.StatusInternalServerError)
		log.Println("    could not save file: " + err.Error())
		return
	}

	// Implicitly means code 200
	w.Write([]byte(link))
}

// Generates a valid link to uploadFile with the specified file extension.
// Returns the link or an error in case of failure.
// Does not close the passed file pointer.
func generateLink(handler *uploadHandler, uploadFile *multipart.File, fileExtension string) (string, error) {
	// Find an unused file name
	var fullFileName string
	var savePath string
	for {
		fileStem := createRandomFileName(handler.config.LinkLength)
		fullFileName = fileStem + fileExtension
		savePath = handler.config.FileDir + fullFileName

		if !fileExists(savePath) {
			break
		}
	}

	link := handler.config.LinkPrefix + fullFileName

	err := saveFile(uploadFile, savePath)
	if err != nil {
		return "", err
	}

	return link, nil
}

func saveFile(data *multipart.File, name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, *data)
	if err != nil {
		return err
	}

	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	return !errors.Is(err, os.ErrNotExist)
}

func createRandomFileName(length int) string {
	chars := make([]byte, length)

	for i := 0; i < length; i++ {
		index := rand.Intn(len(allowedChars))
		chars[i] = allowedChars[index]
	}

	return string(chars)
}

func splitFileName(name string) (string, string) {
	extIndex := strings.LastIndex(name, ".")

	if extIndex == -1 {
		// No dot at all
		return name, ""
	}

	return name[:extIndex], name[extIndex:]
}
