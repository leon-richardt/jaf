package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/leon-richardt/jaf/exifscrubber"
)

const allowedChars = "0123456789ABCDEFGHIJKLMNOPQRTSUVWXYZabcdefghijklmnopqrstuvwxyz"

var config Config

type parameters struct {
	configFile string
}

func parseParams() *parameters {
	configFile := flag.String("configFile", "jaf.conf", "path to config file")
	flag.Parse()

	retval := &parameters{}
	retval.configFile = *configFile
	return retval
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetPrefix("jaf > ")

	params := parseParams()

	// Read config
	config, err := ConfigFromFile(params.configFile)
	if err != nil {
		log.Fatalf("could not parse config file: %s\n", err.Error())
	}

	handler := uploadHandler{
		config: config,
	}

	if config.ScrubExif {
		scrubber := exifscrubber.NewExifScrubber(config.ExifAllowedIds, config.ExifAllowedPaths)
		handler.exifScrubber = &scrubber
	}

	// Start server
	uploadServer := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         fmt.Sprintf(":%d", config.Port),
	}

	log.Printf("starting jaf on port %d\n", config.Port)
	http.Handle("/upload", &handler)
	uploadServer.ListenAndServe()
}
