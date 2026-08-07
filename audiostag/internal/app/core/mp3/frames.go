package mp3

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	_ "golang.org/x/image/webp"

	"github.com/cabbagekobe/tunetag"
	"github.com/cabbagekobe/tunetag/id3v2"

	"github.com/michaeluuong/utilize/filing"
)

func AddAPICURL(mp3 *tunetag.MP3, filePath string) error {
	dir, _ := filepath.Split(filePath)

	coverSource := taggerConfig.TagAttr.CoverSource
	var coverFile string
	var err error
	if !CoverExists(dir) {
		if coverSource != "" {
			if matched, _ := regexp.MatchString("^https?://", coverSource); matched { // URL
				fmt.Printf("GETting image from URL, coverSource: %s\n", coverSource)
				if !CoverExists(dir) {
					coverFile, err = DownloadPicture(taggerConfig.TagAttr.CoverSource, dir, taggerConfig.CoverFilenameNoExt+"69")
					if err != nil {
						slog.Error("DownloadPicture()|could not GET cover art", "coverFile", coverFile, "err", err)
						return err

					}

				}

			} else if filing.Exists(coverSource) { // Local file
				fmt.Printf("LOCAL FILE\n")
				ext := filepath.Ext(coverSource)
				coverFile = filepath.Join(dir, taggerConfig.CoverFilenameNoExt+"68"+ext)
				err = filing.CopyFile(coverSource, coverFile)
				if err != nil {
					slog.Error("filing.CopyFile()|could not copy file", "coverSource", coverSource, "coverFile", coverFile)
					return err

				}

			}

		}

		if coverFile == "" && !taggerConfig.TagAttr.NoBrainz { // download
			fmt.Printf("DOWNLOAD\n")
			artist, album := mp3.V2.Artist(), mp3.V2.Album()
			coverFile, err = CoverArt(artist, album, dir, taggerConfig.CoverFilenameNoExt)
			if err != nil {
				taggerConfig.TagAttr.NoBrainz = true
				slog.Error("CoverArt()|could not GET cover art", "dir", dir, "err", err)
				return err

			}

		}

	}

	if coverFile == "" {
		coverFiles := filing.LsEntryName(dir, taggerConfig.ArtworkRegex(), "add_dir")
		fmt.Printf("coverFiles: %v\n", coverFiles)
		if len(coverFiles) > 0 {
			coverFile = coverFiles[0]

		}

	}

	file, err := os.Open(coverFile)
	if err != nil {
		slog.Error("os.Open()|could not open file", "coverSource", coverSource, "err", err)
		return err

	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		slog.Error("image.Decode()|could not decode image", "file", file, "format", format, "err", err)
		return err

	}
	fmt.Printf("format: %v\n", format)

	var artworkData bytes.Buffer
	jpeg.Encode(&artworkData, img, &jpeg.Options{Quality: 100})

	pictureFrame := &id3v2.PictureFrame{
		Encoding:    id3v2.EncUTF8,
		MIME:        "image/" + format,
		PictureType: uint8(tunetag.PictureCoverFront),
		Description: "Cover",
		Data:        artworkData.Bytes(),
	}

	//slog.Info("picture file type found adding frame", "filename", filename, "fileType", fileType)
	pictureFrames := mp3.V2.PictureFrames()
	if len(pictureFrames) > 0 {
		slog.Debug("mp3.V2.PictureFrames()|cover art frame exists removing")
		mp3.V2.RemoveFrames("APIC")

	}
	mp3.V2.AddFrame(pictureFrame)

	return nil

}

func AddAPICData(mp3 *tunetag.MP3, data []byte) error {
	pictureFrames := mp3.V2.PictureFrames()
	if data != nil && len(pictureFrames) == 0 {
		mimeType := http.DetectContentType(data)

		pictureFrame := &id3v2.PictureFrame{
			Encoding:    id3v2.EncUTF8,
			MIME:        mimeType,
			PictureType: uint8(tunetag.PictureCoverFront),
			Description: "Cover",
			Data:        data,
		}

		mp3.V2.AddFrame(pictureFrame)

	}

	return nil

}

// AddAPIC adds an existing cover art image to the APIC frame of the mp3 file.
func AddAPIC(mp3 *tunetag.MP3, filePath string, overwrite bool) error {
	dir, filename := filepath.Split(filePath)

	pictureFrames := mp3.V2.PictureFrames()
	if len(pictureFrames) > 0 && !overwrite {
		slog.Debug("input|picture exists", "len(pictureFrames)", len(pictureFrames), "overwrite", overwrite)
		return nil

	}

	artworkFiles, _ := filing.Ls(dir, taggerConfig.ArtworkRegex())

	pictureFrame := &id3v2.PictureFrame{}
	if len(artworkFiles) > 0 {
		if len(pictureFrames) > 0 {
			//slog.Debug("mp3.V2.PictureFrames()|cover art frame exists removing", "pictureFrames", pictureFrames)
			slog.Debug("mp3.V2.PictureFrames()|cover art frame exists removing", "filename", filename)
			mp3.V2.RemoveFrames("APIC")

		}

		artworkFile := filepath.Join(dir, artworkFiles[0].Name())
		file, err := os.Open(artworkFile)
		if err != nil {
			slog.Error("os.Open()|could not open file", "artworkFile", artworkFile, "err", err)
			return err

		}
		defer file.Close()

		img, format, err := image.Decode(file)
		if err != nil {
			slog.Error("image.Decode()|could not decode image", "file", file, "format", format, "err", err)
			return err

		}

		var artworkData bytes.Buffer
		jpeg.Encode(&artworkData, img, &jpeg.Options{Quality: 100})

		pictureFrame = &id3v2.PictureFrame{
			Encoding:    id3v2.EncUTF8,
			MIME:        "image/" + format,
			PictureType: uint8(tunetag.PictureCoverFront),
			Description: "Cover",
			Data:        artworkData.Bytes(),
		}

		slog.Debug("AddAPIC()|picture file found adding frame", "filename", filename, "format", format)
		mp3.V2.AddFrame(pictureFrame)

	}

	return nil

}

func addCommentFrame(mp3 *tunetag.MP3, comment string) {
	mp3.V2.RemoveFrames("COMM")
	mp3.V2.AddFrame(&id3v2.CommentFrame{
		Encoding: id3v2.EncUTF8,
		Language: "eng",
		Text:     comment,
	})

}

func addGenericFrame(mp3 *tunetag.MP3, frameID, body string) {
	if body != "" {
		mp3.V2.RemoveFrames(frameID)
		mp3.V2.AddFrame(&id3v2.GenericFrame{
			FrameID: frameID,
			//Body:        []byte("\x00" + body + "\x00"),
			Body:        []byte(body),
			StatusFlags: byte(0),
			FormatFlags: byte(0),
		})

	}

}

func addPrivFrame(mp3 *tunetag.MP3, frameID, data, owner string) {
	if owner != "" || data != "" {
		mp3.V2.RemoveFrames(frameID)
		mp3.V2.AddFrame(&id3v2.PrivFrame{
			Owner: owner,
			//Data:  []byte("\x00" + data + "\x00"),
			Data: []byte(data),
		})

	}

}

func addUFIDFrame(mp3 *tunetag.MP3, frameID, identifier, owner string) {
	if owner != "" || identifier != "" {
		mp3.V2.RemoveFrames(frameID)
		mp3.V2.AddFrame(&id3v2.UFIDFrame{
			Owner: owner,
			//Identifier: []byte("\x00" + identifier + "\x00"),
			Identifier: []byte(identifier),
		})

	}
}

func addURLFrame(mp3 *tunetag.MP3, frameID, url string) {
	if url != "" {
		mp3.V2.RemoveFrames(frameID)
		mp3.V2.AddFrame(&id3v2.URLFrame{
			FrameID: frameID,
			URL:     url,
		})

	}

}

func addUserTextFrame(mp3 *tunetag.MP3, value, description string) {
	if value != "" {
		//mp3.V2.RemoveFrames("TXXX")
		mp3.V2.AddFrame(&id3v2.UserTextFrame{
			Encoding:    id3v2.EncUTF8,
			Description: description,
			Value:       value,
		})

	}

}

// addUnsychronisedLyricsFrame
func addUnsynchronisedLyricsFrame(mp3 *tunetag.MP3, frameID, text, description string) {
	if text != "" {
		mp3.V2.RemoveFrames(frameID)
		mp3.V2.AddFrame(&id3v2.UnsynchronisedLyricsFrame{
			Encoding:    id3v2.EncUTF8,
			Language:    "eng",
			Description: description,
			Text:        text,
		})

	}

}

func addUserURLFrame(mp3 *tunetag.MP3, frameID, url, description string) {
	if description != "" || url != "" {
		mp3.V2.RemoveFrames(frameID)
		mp3.V2.AddFrame(&id3v2.UserURLFrame{
			Encoding:    id3v2.EncUTF8,
			Description: description,
			URL:         url,
		})

	}

}
