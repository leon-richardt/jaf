package extdetect

import (
	"os"
	"testing"
)

func TestDetectedExtensions(t *testing.T) {
	const fixturePath = "../fixtures/gps.png"

	type tType struct {
		name           string
		fileData       []byte
		expectedOutput string
	}

	pngFile, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Could not open \"%s\" which is required for the test. Error: %s", fixturePath,
			err)
	}

	tests := []tType{
		{ // extension is detected correctly from file when not specified explicitly
			name:           "foo",
			fileData:       pngFile,
			expectedOutput: ".png",
		},
		{
			name:           "foo.txt",
			expectedOutput: ".txt",
		},
		{ // simple extension that's the last part of a known combination is detected correctly
			name:           "foo.gz",
			expectedOutput: ".gz",
		},
		{ // simple extension that's the first part of a known combination is detected correctly
			name:           "foo.tar",
			expectedOutput: ".tar",
		},
		{ // combined extension is detected correctly
			name:           "foo.tar.gz",
			expectedOutput: ".tar.gz",
		},
		{
			name:           "foo.tar.xz",
			expectedOutput: ".tar.xz",
		},
		{ // combined extension that is NOT known only returns the last part
			name:           "foo.jpg.zip",
			expectedOutput: ".zip",
		},
		{ // combined extension is detected correctly even with many "." in the name
			name:           "foo.jpg.zip.tar.gz",
			expectedOutput: ".tar.gz",
		},
	}

	for _, test := range tests {
		output := BuildFileExtension(test.fileData, test.name)
		if output != test.expectedOutput {
			t.Fatalf("got output '%s', expected '%s'", output, test.expectedOutput)
		}
	}
}
