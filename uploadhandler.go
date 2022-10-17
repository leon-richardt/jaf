package main

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/leon-richardt/jaf/exifscrubber"
	"github.com/leon-richardt/jaf/extdetect"
)

type uploadHandler struct {
	config       *Config
	exifScrubber *exifscrubber.ExifScrubber
}

func (handler *uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	uploadFile, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "could not read uploaded file: "+err.Error(), http.StatusBadRequest)
		log.Println("    could not read uploaded file: " + err.Error())
		return
	}

	fileData, err := io.ReadAll(uploadFile)
	uploadFile.Close()
	if err != nil {
		http.Error(w, "could not read attached file: "+err.Error(), http.StatusInternalServerError)
		log.Println("    could not read attached file: " + err.Error())
		return
	}

	// Scrub EXIF, if requested and detectable by us
	if handler.config.ScrubExif {
		scrubbedData, err := handler.exifScrubber.ScrubExif(fileData[:])

		if err == nil {
			// If scrubbing was successful, update what to write to file
			fileData = scrubbedData
		} else {
			// Unknown file types (not PNG or JPEG) are allowed to contain EXIF, as we don't know
			// how to handle them. Handling of other errors depends on configuration.
			if err != exifscrubber.ErrUnknownFileType {
				if handler.config.ExifAbortOnError {
					log.Printf("could not scrub EXIF from file, aborting upload: %s", err.Error())
					http.Error(
						w,
						"could not scrub EXIF from file: "+err.Error(),
						http.StatusInternalServerError,
					)
					return
				}

				// An error occured but we are configured to proceed with the upload anyway
				log.Printf(
					"could not scrub EXIF from file but proceeding with upload as configured: %s",
					err.Error(),
				)
			}
		}
	}

	link, err := generateLink(handler, fileData[:], header.Filename)
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
func generateLink(handler *uploadHandler, fileData []byte, fileName string) (string, error) {
	ext := extdetect.BuildFileExtension(fileData, fileName)

	// Find an unused file name
	var fullFileName string
	var savePath string
	for {
		fileStem := createRandomFileName(handler.config.LinkLength)
		fullFileName = fileStem + ext
		savePath = handler.config.FileDir + fullFileName

		if !fileExists(savePath) {
			break
		}
	}

	link := handler.config.LinkPrefix + fullFileName

	err := saveFile(fileData[:], savePath)
	if err != nil {
		return "", err
	}

	return link, nil
}

func saveFile(fileData []byte, name string) error {
	err := os.WriteFile(name, fileData, 0o644)
	return err
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)

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
