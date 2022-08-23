package exifscrubber

import (
	"io/ioutil"
	"log"
	"testing"

	"golang.org/x/exp/slices"

	exif "github.com/dsoprea/go-exif/v3"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	pis "github.com/dsoprea/go-png-image-structure/v2"
)

func TestJpgFromFile(t *testing.T) {
	// Read and strip original file
	buf, err := ioutil.ReadFile("../fixtures/gps.jpg")
	if err != nil {
		t.Errorf("could not open file")
	}

	includeTagIds := []uint16{0x9209} // ID of "Flash" tag
	includedPaths := []string{
		"IFD/Orientation",
		"IFD/GPSInfo/GPSTimeStamp",
		"IFD/GPSInfo/GPSDateStamp",
	}

	scrubber := NewExifScrubber(includeTagIds[:], includedPaths[:])

	updatedBuf, err := scrubber.ScrubExif(buf)
	if err != nil {
		log.Println(err)
	}

	// Check whether updated file only contains the specified paths
	intfc, err := jis.NewJpegMediaParser().ParseBytes(updatedBuf)
	if err != nil {
		t.Errorf(err.Error())
	}

	sl := intfc.(*jis.SegmentList)
	rootIfd, _, err := sl.Exif()

	visitor := func(ifd *exif.Ifd, ite *exif.IfdTagEntry) error {
		tagId := ite.TagId()
		tagPath := ite.IfdPath() + "/" + ite.TagName()

		tagContained := slices.Contains(includeTagIds, tagId)
		pathContained := slices.Contains(includedPaths, tagPath)
		contained := tagContained || pathContained

		if !contained {
			t.Errorf(
				"tag %s (%d) included in EXIF although it hasn't been specified",
				tagPath,
				tagId,
			)
		}

		return nil
	}

	rootIfd.EnumerateTagsRecursively(visitor)
}

func TestPngFromFile(t *testing.T) {
	buf, err := ioutil.ReadFile("../fixtures/gps.png")
	if err != nil {
		t.Errorf("could not open file")
	}

	includeTagIds := []uint16{0x9209} // ID of "Flash" tag
	includedPaths := []string{
		"IFD/Orientation",
		"IFD/GPSInfo/GPSTimeStamp",
		"IFD/GPSInfo/GPSDateStamp",
	}

	scrubber := NewExifScrubber(includeTagIds[:], includedPaths[:])

	updatedBuf, err := scrubber.ScrubExif(buf)
	if err != nil {
		log.Println(err)
	}

	// Check whether updated file only contains the specified paths
	intfc, err := pis.NewPngMediaParser().ParseBytes(updatedBuf)
	if err != nil {
		t.Errorf(err.Error())
	}

	cs := intfc.(*pis.ChunkSlice)
	rootIfd, _, err := cs.Exif()

	visitor := func(ifd *exif.Ifd, ite *exif.IfdTagEntry) error {
		tagId := ite.TagId()
		tagPath := ite.IfdPath() + "/" + ite.TagName()

		tagContained := slices.Contains(includeTagIds, tagId)
		pathContained := slices.Contains(includedPaths, tagPath)
		contained := tagContained || pathContained

		if !contained {
			t.Errorf(
				"tag %s (%d) included in EXIF although it hasn't been specified",
				tagPath,
				tagId,
			)
		}

		return nil
	}

	rootIfd.EnumerateTagsRecursively(visitor)
}
