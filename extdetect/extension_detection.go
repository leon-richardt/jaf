package extdetect

import (
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

var knownCombinations []string = []string{
	".tar.gz",
	".tar.xz",
}

func BuildFileExtension(fileData []byte, name string) string {
	// First, check whether any file ending has been specified manually
	curExtIdx := strings.LastIndex(name, ".")

	if curExtIdx == -1 {
		// No file ending specified in name, use MIME type detection
		return mimetype.Detect(fileData).Extension()
	}

	// Otherwise, some file extension was manually specified and we will use that. First, check
	// whether this is an "easy" case of file extension, i.e., a name where there is only one "."
	// character and we can treat what's after it as the file extension.
	nextExtIdx := strings.LastIndex(name[:curExtIdx], ".")
	if nextExtIdx == -1 {
		// Just one ".", so an easy case
		return name[curExtIdx:]
	}

	// There are multiple "." in the name. Look for known extension combinations (e.g., ".tar.gz",
	// ".tar.xz") and use that if found.
	// XXX: This could be done more efficiently (at least in theory) with some suffix tree structure
	//      but for the few known combinations we have, it would likely be slower on real-world
	//      computer architectures.
	stillBuilding := true
	for stillBuilding {
		stillBuilding = false
		for _, comb := range knownCombinations {
			if !strings.HasPrefix(comb, name[nextExtIdx:]) {
				continue
			}

			stillBuilding = true
			curExtIdx = nextExtIdx
			nextExtIdx = strings.LastIndex(name[:curExtIdx], ".")
			if nextExtIdx == -1 {
				// No more extension candidates -> return current state of the builder
				return name[curExtIdx:]
			}
		}
	}

	return name[curExtIdx:]
}
