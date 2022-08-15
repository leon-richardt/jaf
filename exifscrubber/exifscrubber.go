package exifscrubber

import (
	"bytes"
	"errors"
	"fmt"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	pis "github.com/dsoprea/go-png-image-structure/v2"
)

var ErrUnknownFileType = errors.New("can't scrub EXIF for this file type")

type ExifScrubber struct {
	includedTagIds   []uint16
	includedTagPaths []string
}

func NewExifScrubber(includedTagIds []uint16, includedTagPaths []string) ExifScrubber {
	return ExifScrubber{
		includedTagIds:   includedTagIds,
		includedTagPaths: includedTagPaths,
	}
}

func (scrubber *ExifScrubber) ScrubExif(fileData []byte) ([]byte, error) {
	// Try scrubbing using JPEG package
	jpegParser := jis.NewJpegMediaParser()
	if jpegParser.LooksLikeFormat(fileData) {
		intfc, err := jpegParser.ParseBytes(fileData)
		if err != nil {
			return nil, err
		}

		segmentList := intfc.(*jis.SegmentList)
		rootIfd, _, err := segmentList.Exif()
		if err != nil {
			if err == exif.ErrNoExif {
				// Incoming data contained no EXIF in the first place so we can return the original
				return fileData, nil
			}

			return nil, err
		}

		filteredIb, err := scrubber.filteringIfdBuilder(rootIfd)
		if err != nil {
			return nil, err
		}
		segmentList.SetExif(filteredIb)

		b := new(bytes.Buffer)
		err = segmentList.Write(b)
		if err != nil {
			return nil, err
		}

		return b.Bytes(), nil
	}

	// Try scrubbing using PNG package
	pngParser := pis.NewPngMediaParser()
	if pngParser.LooksLikeFormat(fileData) {
		intfc, err := pngParser.ParseBytes(fileData)
		if err != nil {
			return nil, err
		}

		chunks := intfc.(*pis.ChunkSlice)
		rootIfd, _, err := chunks.Exif()
		if err != nil {
			if err == exif.ErrNoExif {
				// Incoming data contained no EXIF in the first place so we can return the original
				return fileData, nil
			}

			return nil, err
		}

		filteredIb, err := scrubber.filteringIfdBuilder(rootIfd)
		if err != nil {
			return nil, err
		}
		chunks.SetExif(filteredIb)

		b := new(bytes.Buffer)
		err = chunks.WriteTo(b)
		if err != nil {
			return nil, err
		}

		return b.Bytes(), nil
	}

	// Don't know how to handle other file formats, so we let the caller decide how to continue
	return nil, ErrUnknownFileType
}

// Check whether the tag represented by `tag` is included in the path or tag ID list
func (scrubber *ExifScrubber) isTagAllowed(tag *exif.IfdTagEntry) bool {
	// Check via IDs first (faster than string comparisons)
	for _, includedId := range scrubber.includedTagIds {
		if includedId == tag.TagId() {
			return true
		}
	}

	// If no IDs matched, also check IFD tag paths for inclusion
	tagPath := fmt.Sprintf("%s/%s", tag.IfdPath(), tag.TagName())

	for _, includedPath := range scrubber.includedTagPaths {
		if includedPath == tagPath {
			return true
		}
	}

	return false
}

// This method follows the implementation of exif.NewIfdBuilderFromExistingChain()
func (scrubber *ExifScrubber) filteringIfdBuilder(rootIfd *exif.Ifd) (
	firstIb *exif.IfdBuilder,
	err error,
) {
	var lastIb *exif.IfdBuilder
	i := 0
	for thisExistingIfd := rootIfd; thisExistingIfd != nil; thisExistingIfd = thisExistingIfd.NextIfd() {
		// This only works when no non-standard mappings are used
		ifdMapping, err := exifcommon.NewIfdMappingWithStandard()
		if err != nil {
			return nil, err
		}

		// This only works when no non-standard tags are used
		tagIndex := exif.NewTagIndex()
		err = exif.LoadStandardTags(tagIndex)
		if err != nil {
			return nil, err
		}

		newIb := exif.NewIfdBuilder(
			ifdMapping,
			tagIndex,
			thisExistingIfd.IfdIdentity(),
			thisExistingIfd.ByteOrder(),
		)

		if firstIb == nil {
			firstIb = newIb
		} else {
			lastIb.SetNextIb(newIb)
		}

		err = scrubber.filteredAddTagsFromExisting(newIb, thisExistingIfd)
		if err != nil {
			return nil, err
		}

		lastIb = newIb
		i++
	}

	return firstIb, nil
}

// This method follows the implementation of exif.IfdBuilder.AddTagsFromExisting()
func (scrubber *ExifScrubber) filteredAddTagsFromExisting(
	ib *exif.IfdBuilder,
	ifd *exif.Ifd,
) (err error) {
	for i, ite := range ifd.Entries() {
		if ite.IsThumbnailOffset() == true || ite.IsThumbnailSize() {
			// These will be added on-the-fly when we encode.
			continue
		}

		var bt *exif.BuilderTag
		if ite.ChildIfdPath() != "" {
			// If we want to add an IFD tag, we'll have to build it first and
			// *then* add it via a different method.

			// Figure out which of the child-IFDs that are associated with
			// this IFD represents this specific child IFD.

			var childIfd *exif.Ifd
			for _, thisChildIfd := range ifd.Children() {
				if thisChildIfd.ParentTagIndex() != i {
					continue
				} else if thisChildIfd.IfdIdentity().TagId() != 0xffff &&
					thisChildIfd.IfdIdentity().TagId() != ite.TagId() {
					fmt.Printf(
						"child-IFD tag is not correct: TAG-POSITION=(%d) ITE=%s CHILD-IFD=%s\n",
						thisChildIfd.ParentTagIndex(),
						ite,
						thisChildIfd,
					)
				}

				childIfd = thisChildIfd
				break
			}

			if childIfd == nil {
				childTagIds := make([]string, len(ifd.Children()))
				for j, childIfd := range ifd.Children() {
					childTagIds[j] = fmt.Sprintf(
						"0x%04x (parent tag-position %d)",
						childIfd.IfdIdentity().TagId(),
						childIfd.ParentTagIndex(),
					)
				}

				fmt.Printf(
					"could not find child IFD for child ITE: IFD-PATH=[%s] TAG-ID=(0x%04x) "+
						"CURRENT-TAG-POSITION=(%d) CHILDREN=%v\n",
					ite.IfdPath(),
					ite.TagId(),
					i,
					childTagIds,
				)
			}

			childIb, err := scrubber.filteringIfdBuilder(childIfd)
			if err != nil {
				return err
			}

			bt = ib.NewBuilderTagFromBuilder(childIb)
		} else {
			// Non-IFD tag.
			isAllowed := scrubber.isTagAllowed(ite)
			if !isAllowed {
				continue
			}

			rawBytes, err := ite.GetRawBytes()
			if err != nil {
				return err
			}

			value := exif.NewIfdBuilderTagValueFromBytes(rawBytes)
			bt = exif.NewBuilderTag(
				ifd.IfdIdentity().UnindexedString(),
				ite.TagId(),
				ite.TagType(),
				value,
				ifd.ByteOrder(),
			)
		}

		if bt.Value().IsBytes() {
			err := ib.Add(bt)
			if err != nil {
				return err
			}
		} else if bt.Value().IsIb() {
			err := ib.AddChildIb(bt.Value().Ib())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
