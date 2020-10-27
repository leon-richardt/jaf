package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	commentPrefix = "#"
)

type Config struct {
	Port       int
	LinkPrefix string
	FileDir    string
	LinkLength int
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

	retval := &Config{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, commentPrefix) {
			// Skip comments
			continue
		}

		tokens := strings.Split(line, ": ")
		if len(tokens) != 2 {
			log.Printf("unexpected line: \"%s\", ignoring\n", line)
			continue
		}

		key, val := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])

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
		default:
			log.Printf("unexpected key: \"%s\", ignoring\n", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return retval, nil
}
