package main

import (
	"testing"
)

func assertEqual[S comparable](have S, want S, t *testing.T) {
	if have != want {
		t.Error("have:", have, ", want:", want, "\n")
	}
}

func TestConfigFromFile(t *testing.T) {
	config, err := ConfigFromFile("example.conf")
	if err != nil {
		panic(err)
	}

	assertEqual(config.Port, 4711, t)
	assertEqual(config.LinkPrefix, "https://jaf.example.com/", t)
	assertEqual(config.FileDir, "/var/www/jaf/", t)
	assertEqual(config.LinkLength, 5, t)
}
