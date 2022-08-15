package main

import (
	"testing"
)

func assertEqual[S comparable](have S, want S, t *testing.T) {
	if have != want {
		t.Error("have:", have, ", want:", want)
	}
}

func assertEqualSlice[S comparable](have []S, want []S, t *testing.T) {
	if len(have) != len(want) {
		t.Error("lengths differ! have:", len(have), ", want:", len(want))
		return
	}

	for i := range want {
		if have[i] != want[i] {
			t.Error("slices differ at position", i, ":", have[i], "!=", want[i])
			return
		}
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
	assertEqual(config.ScrubExif, true, t)
	assertEqualSlice(config.ExifAllowedIds, []uint16{0x0112, 274}, t)
	assertEqualSlice(config.ExifAllowedPaths, []string{"IFD/Orientation"}, t)
	assertEqual(config.ExifAbortOnError, true, t)
}
