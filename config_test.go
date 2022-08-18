package main

import (
	"testing"
)

func assertEqualInt(have int, want int, t *testing.T) {
	if have != want {
		t.Errorf("have: %d, want: %d\n", have, want)
	}
}

func assertEqualString(have string, want string, t *testing.T) {
	if have != want {
		t.Errorf("have: %s, want: %s\n", have, want)
	}
}

func TestConfigFromFile(t *testing.T) {
	config, err := ConfigFromFile("example.conf")
	if err != nil {
		panic(err)
	}

	assertEqualInt(config.Port, 4711, t)
	assertEqualString(config.LinkPrefix, "https://jaf.example.com/", t)
	assertEqualString(config.FileDir, "/var/www/jaf/", t)
	assertEqualInt(config.LinkLength, 5, t)
}
