package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
)

const (
	commentPrefix = "#"
)

type Config struct {
	Port             int
	LinkPrefix       string
	FileDir          string
	LinkLength       int
	ScrubExif        bool
	ExifAllowedIds   []uint16
	ExifAllowedPaths []string
	ExifAbortOnError bool
}

func ConfigFromFile(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	oldPrefix := log.Prefix()
	defer log.SetPrefix(oldPrefix)

	log.SetPrefix("config.FromFile > ")

	retval := &Config{
		Port:             4711,
		LinkPrefix:       "https://jaf.example.com/",
		FileDir:          "/var/www/jaf/",
		LinkLength:       5,
		ScrubExif:        true,
		ExifAllowedIds:   []uint16{},
		ExifAllowedPaths: []string{},
		ExifAbortOnError: true,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, commentPrefix) {
			// Skip comments
			continue
		}

		key, val, found := strings.Cut(line, ":")

		if !found {
			log.Printf("unexpected line: \"%s\", ignoring\n", line)
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch key {
		case "Port":
			parsed, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}

			retval.Port = parsed
		case "LinkPrefix":
			retval.LinkPrefix = val
		case "FileDir":
			retval.FileDir = val
		case "LinkLength":
			parsed, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}

			retval.LinkLength = parsed
		case "ScrubExif":
			parsed, err := strconv.ParseBool(val)
			if err != nil {
				return nil, err
			}

			retval.ScrubExif = parsed
		case "ExifAllowedIds":
			if val == "" {
				// No IDs specified at all
				break
			}

			stringIds := strings.Split(val, " ")

			parsedIds := make([]uint16, 0, len(stringIds))
			for _, stringId := range stringIds {
				var parsed uint64
				var err error

				if strings.HasPrefix(stringId, "0x") {
					// Parse as a hexadecimal number
					hexStringId := strings.Replace(stringId, "0x", "", 1)
					parsed, err = strconv.ParseUint(hexStringId, 16, 16)
				} else {
					// Parse as a decimal number
					parsed, err = strconv.ParseUint(stringId, 10, 16)
				}

				if err != nil {
					log.Printf(
						"Could not parse ID from: \"%s\", ignoring. Error: %s\n",
						stringId,
						err,
					)
					continue
				}

				parsedIds = append(parsedIds, uint16(parsed))
			}

			retval.ExifAllowedIds = parsedIds
		case "ExifAllowedPaths":
			if val == "" {
				// No paths specified at all
				break
			}

			paths := strings.Split(val, " ")
			retval.ExifAllowedPaths = paths
		case "ExifAbortOnError":
			parsed, err := strconv.ParseBool(val)
			if err != nil {
				return nil, err
			}

			retval.ExifAbortOnError = parsed
		default:
			return nil, errors.Errorf("unexpected config key: \"%s\"", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return retval, nil
}
