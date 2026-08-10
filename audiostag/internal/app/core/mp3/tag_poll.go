package mp3

import (
	"archive/tar"
	"bytes"
	"cmp"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/cabbagekobe/tunetag/id3v1"
	"github.com/cabbagekobe/tunetag/id3v2"
	"github.com/dhowden/tag"

	"github.com/michaeluuong/audios/audiostag/internal/app/core/mp3/config"
	"github.com/michaeluuong/audios/audiostag/internal/pkg/aud_io"
	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/reflections"
	"github.com/michaeluuong/utilize/stringy"
)

/*
	var headerRow = []string{
		"Title",
		"Artist",
		"Album",
		"Comment",
		"Year",
		"Track",
		"Genre",
		"Format",
		"FileType",
		"Disc",
		"Picture",
	}
*/

func PrintTags(dir string, outFilenameOpt ...string) error {
	slog.Info("starting", "dir", dir, "outFilenameOpt", outFilenameOpt)
	if len(outFilenameOpt) == 1 && outFilenameOpt[0] == "" {
		var emptyStringSlice []string
		outFilenameOpt = emptyStringSlice

	}

	data := [][]string{}
	var img image.Image
	mp3Files := filing.LsEntryName(dir, ".*\\.mp3", "add_dir")
	cImg := aud_io.CoverImage{}
	for i, mp3File := range mp3Files {
		fmt.Printf("mp3File: %s\n", mp3File)
		file, err := os.Open(mp3File)
		if err != nil {
			return err

		}
		defer file.Close()

		var getImage bool
		if i == 0 && img == nil {
			getImage = true

		}

		filename := filepath.Base(mp3File)
		if imgCheck, err := readMP3Tags(file, filename, &data, getImage); err != nil {
			slog.Error("readMP3Tags()|error reading tags", "err", err)
			return err

		} else if imgCheck != nil {
			img = imgCheck
			cImg.Set("embedded", img)

		}

	}

	if len(data) == 0 {
		return fmt.Errorf("There are no files to process in dir: %s", dir)

	}

	sort2DSliceByTrackAndDisc(data)

	var writer io.Writer = os.Stdout
	if len(outFilenameOpt) > 0 && outFilenameOpt[0] != "" {
		//if outFilenameOpt[0] == "window" {
		if slices.Contains(outFilenameOpt, "window") {
			showActions, hasCompleted := false, false
			if slices.Contains(outFilenameOpt, "showActions") {
				showActions = true

			}

			if slices.Contains(outFilenameOpt, "completed") {
				hasCompleted = true

			}

			artFiles := filing.LsEntryName(dir, ".*\\.(jpg|png|gif)", "add_dir")
			if img == nil && len(artFiles) > 0 {
				imgFile, err := os.Open(artFiles[0])
				if err != nil {
					return err

				}
				defer imgFile.Close()
				img, _, err = image.Decode(imgFile)
				if err != nil {
					return err

				}

				cImg.Set("embedded", img)

			}

			winman := aud_io.GetWinmanInstance()
			winman.TagTableShow(dir, data, cImg, showActions, hasCompleted)

			return nil

		} else {
			oFile, err := os.OpenFile(outFilenameOpt[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, filing.FilePerm)
			if err != nil {
				slog.Error("os.OpenFile()|could not open file", "outFilename", outFilenameOpt[0])

			}
			defer oFile.Close()
			writer = oFile

		}

	}

	stringy.PrintDataSlice(dir, data, writer)

	return nil

}

func PrintArchiveTags(ctx context.Context, filePath string, outFilenameOpt ...string) error {
	if len(outFilenameOpt) > 0 && outFilenameOpt[0] == "" {
		var emptyStringSlice []string
		outFilenameOpt = emptyStringSlice

	}

	filename := filepath.Base(filePath)
	fmt.Printf("filename: %s\n", filename)
	file, err := os.Open(filePath)
	if err != nil {
		return err

	}
	defer file.Close()

	var reader io.Reader
	ext := filepath.Ext(filePath)
	if ext == ".gz" {
		reader, err = gzip.NewReader(file)
		if err != nil {
			slog.Error("gzip.NewReader()|could not get reader for gzip file", "filePath", filePath, "err", err)
			return err
		}

	} else if ext == ".tar" {
		reader = file

	} else {
		slog.Error("only tar and gzip files are currently supported", "ext", ext)
		return errors.New("only tar and gzip files are currently supported, ext: " + ext)

	}

	data := [][]string{}
	tarReader := tar.NewReader(reader)
	cImg := &aud_io.CoverImage{}
	i := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break

		}
		if err != nil {
			return err

		}

		if header.FileInfo().IsDir() || header.Typeflag != tar.TypeReg || !strings.HasSuffix(strings.ToLower(header.Name), ".mp3") {
			regex := regexp.MustCompile(`.(jpg|png|gif)`)
			if matches := regex.MatchString(header.Name); matches {
				tarImg, _, _ := image.Decode(tarReader)
				if tarImg != nil {
					cImg.Set(filepath.Base(header.Name), tarImg)

				}

			}

			continue

		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tarReader); err != nil {
			continue

		}

		seeker := bytes.NewReader(buf.Bytes())
		filename := filepath.Base(filepath.Clean(header.Name))

		var getImage bool
		if cImg.Image == nil {
			getImage = true

		}

		if imgCheck, err := readMP3Tags(seeker, filename, &data, getImage); err != nil {
			slog.Error("readMP3Tags()|error reading tags", "err", err)
			return err

		} else if imgCheck != nil && cImg.Image == nil {
			cImg.Set("embedded", imgCheck)

		}
		i++

	}
	sort2DSliceByTrackAndDisc(data)

	var writer io.Writer = os.Stdout
	if len(outFilenameOpt) > 0 && outFilenameOpt[0] != "" {
		fmt.Printf("len(outFilennameOpt): %d, outFilenameOpt[0]: %s\n", len(outFilenameOpt), outFilenameOpt[0])
		//ofile, err = os.Create(outFilenameOpt[0])
		if outFilenameOpt[0] == "window" {
			showActions := false
			if len(outFilenameOpt) > 1 && outFilenameOpt[1] == "true" {
				showActions = true

			}

			winman := aud_io.GetWinmanInstance()
			winman.TagTableShow(filePath, data, *cImg, showActions, false)

			return nil

		} else {
			oFile, err := os.OpenFile(outFilenameOpt[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, filing.FilePerm)
			if err != nil {
				slog.Error("os.OpenFile()|could not open file", "outFilename", outFilenameOpt[0])

			}
			defer oFile.Close()
			writer = oFile

		}

	}

	stringy.PrintDataSlice(filename, data, writer)

	return nil

}

func sort2DSliceByTrackAndDisc(data [][]string) {
	trckIndex := slices.Index(data[0], "TRCK")
	discIndex := slices.Index(data[0], "TPOS")
	fileIndex := slices.Index(data[0], "File")

	slices.SortFunc(data[2:], func(a, b []string) int {
		discNumberA, discNumberB := 1, 1
		if discIndex >= 0 {
			discNumberACheck, _ := strconv.Atoi(strings.Split(a[discIndex], "/")[0])
			discNumberA = max(discNumberA, discNumberACheck)

			discNumberBCheck, _ := strconv.Atoi(strings.Split(b[discIndex], "/")[0])
			discNumberB = max(discNumberB, discNumberBCheck)

			// If disc number is not provided try and use the filename
			if discNumberACheck == 0 {
				if matched, _ := regexp.MatchString("^[0-9]{3}", a[fileIndex]); matched {
					fileRunesA := []rune(a[fileIndex])
					discNumberACheck, _ = strconv.Atoi(string(fileRunesA[0:3]))
					discNumberA = max(discNumberA, discNumberACheck)

					fileRunesB := []rune(b[fileIndex])
					discNumberBCheck, _ = strconv.Atoi(string(fileRunesB[0:3]))
					discNumberB = max(discNumberB, discNumberBCheck)

				}

			}

		}

		trackNumA, _ := strconv.Atoi(fmt.Sprintf("%d%02s", discNumberA, strings.Split(a[trckIndex], "/")[0]))
		trackNumB, _ := strconv.Atoi(fmt.Sprintf("%d%02s", discNumberB, strings.Split(b[trckIndex], "/")[0]))

		return cmp.Compare(trackNumA, trackNumB)

	})

}

func PrintArchiveTags2(ctx context.Context, filePath string, outFilenameOpt ...string) (string, error) {
	if len(outFilenameOpt) > 0 && outFilenameOpt[0] == "" {
		var emptyStringSlice []string
		outFilenameOpt = emptyStringSlice

	}

	filename := filepath.Base(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return "", err

	}
	defer file.Close()

	var reader io.Reader
	ext := filepath.Ext(filePath)
	if ext == ".gz" {
		reader, err = gzip.NewReader(file)
		if err != nil {
			slog.Error("gzip.NewReader()|could not get reader for gzip file", "filePath", filePath, "err", err)
			return "", err
		}

	} else if ext == ".tar" {
		reader = file

	} else {
		slog.Error("only tar and gzip files are currently supported", "ext", ext)
		return "", errors.New("only tar and gzip files are currently supported, ext: " + ext)

	}

	data := make(map[int][]string)

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break

		}
		if err != nil {
			return "", err

		}

		if header.FileInfo().IsDir() || header.Typeflag != tar.TypeReg || !strings.HasSuffix(strings.ToLower(header.Name), ".mp3") {
			continue

		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tarReader); err != nil {
			continue

		}

		seeker := bytes.NewReader(buf.Bytes())

		readMP3Tags2(seeker, data)

	}

	var writer io.Writer = os.Stdout
	var stringBuilder strings.Builder
	if len(outFilenameOpt) > 0 && outFilenameOpt[0] != "" {
		//ofile, err = os.Create(outFilenameOpt[0])
		if outFilenameOpt[0] == "window" {
			writer = &stringBuilder

		} else {
			oFile, err := os.OpenFile(outFilenameOpt[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, filing.FilePerm)
			if err != nil {
				slog.Error("os.OpenFile()|could not open file", "outFilename", outFilenameOpt[0])

			}
			defer oFile.Close()
			writer = oFile

		}

	}

	stringy.PrintData(filename, data, writer)

	var output string
	if fmt.Sprintf("%T", writer) == "*strings.Builder" {
		output = stringBuilder.String()

	}

	return output, nil

}

// NOTE: tags that can appear multiple times are keyed by ID (i.e. not description) so the data will be out of order if the IDs aren't consistent
// e.g. track 1: [TXXX=Rip date, TXXX_0=Source] while track 2: [TXXX=Source, TXXX_0=Rip date]
func readMP3Tags(r io.ReadSeeker, filename string, data *[][]string, imageOpt ...bool) (image.Image, error) {
	mp3, err := tag.ReadFrom(r)
	if err != nil {
		slog.Error("error reading tags", "filename", filename, "err", err)
		return nil, err

	}

	var img image.Image
	if len(imageOpt) > 0 && imageOpt[0] {
		if picture := mp3.Picture(); picture != nil {
			if imageBytes := mp3.Picture().Data; imageBytes != nil {
				reader := bytes.NewReader(imageBytes)
				var format string
				img, format, err = image.Decode(reader)
				slog.Debug("image.Decode()|found image", "format", format)
				if err != nil {
					slog.Error("image.Decode()|error decoding image", "err", err)

				}

			}

		}

	}

	raw := mp3.Raw()
	//track, _ := mp3.Track()

	yearID := "TYER"
	if _, ok := raw["TDRC"].(string); ok {
		yearID = "TDRC"

	}

	//Picture{Ext: jpg, MIMEType: image/jpeg, Type: Cover (front), Description: Cover, Data.Size: 70431}

	newModTags := config.NewModTags()
	if len(*data) == 0 || len((*data)[0]) == 0 { // Header row
		if len(*data) == 0 {
			*data = append(*data, []string{})

		}

		var columnIDs = []string{"File", "TIT2", "TPE1", "TALB", "COMM", yearID, "TRCK", "TCON", "TPOS", "APIC", "TFLT", "Format"}

		// Add other columns not already specified
		for id := range raw {
			checkID := id
			if strings.Contains(id, "_") {
				idParts := strings.Split(id, "_")
				checkID = idParts[0]

			}

			if _, ok := newModTags.IsID(checkID); ok && !slices.Contains(columnIDs, id) {
				if (yearID == "TDRC" && id == "TYER") || (yearID == "TYER" && id == "TDRC") {
					continue

				}

				columnIDs = append(columnIDs, id)

			}

		}

		// Create header row
		*data = append(*data, []string{})
		for _, id := range columnIDs {
			(*data)[0] = append((*data)[0], id)

			checkID := id
			if strings.Contains(id, "_") {
				idParts := strings.Split(id, "_")
				checkID = idParts[0]

			}

			value := raw[id]
			_, value2, valueType := GetTagValues(id, value)
			if valueType == "*tag.Comm" && id != "COMM" {
				(*data)[1] = append((*data)[1], value2)

			} else if name, ok := newModTags.IsID(checkID); ok {
				(*data)[1] = append((*data)[1], name)

			} else {
				(*data)[1] = append((*data)[1], id)

			}

		}

	}
	raw["TFLT"] = fmt.Sprintf("%s", mp3.FileType())
	raw["Format"] = fmt.Sprintf("%s", mp3.Format())

	*data = append(*data, []string{})
	curRow := len(*data) - 1
	for _, id := range (*data)[0] {
		if value, ok := raw[id]; ok || id == "File" {
			value1, value2, _ := GetTagValues(id, value)
			if id == "File" {
				(*data)[curRow] = append((*data)[curRow], filename)

			} else if id == "PRIV" {
				(*data)[curRow] = append((*data)[curRow], value2)

			} else {
				(*data)[curRow] = append((*data)[curRow], value1)

			}

		} else {
			(*data)[curRow] = append((*data)[curRow], "")

		}

	}

	return img, nil

}

func readMP3Tags2(r io.ReadSeeker, data map[int][]string) error {
	mp3, err := tag.ReadFrom(r)
	if err != nil {
		slog.Error("error reading tags", "err", err)
		return err

	}

	raw := mp3.Raw()
	track, _ := mp3.Track()

	yearID := "TYER"
	if _, ok := raw["TDRC"].(string); ok {
		yearID = "TDRC"

	}

	//Picture{Ext: jpg, MIMEType: image/jpeg, Type: Cover (front), Description: Cover, Data.Size: 70431}
	newModTags := config.NewModTags()
	if len(data[0]) == 0 { // Header row
		var columnIDs = []string{"TIT2", "TPE1", "TALB", "COMM", yearID, "TRCK", "TCON", "TPOS", "APIC", "TFLT", "Format"}

		// Add other columns to columnIDs that don't already exist
		for id := range raw {
			checkID := id
			if strings.Contains(id, "_") {
				idParts := strings.Split(id, "_")
				checkID = idParts[0]

			}

			if _, ok := newModTags.IsID(checkID); ok && !slices.Contains(columnIDs, id) {
				if (yearID == "TDRC" && id == "TYER") || (yearID == "TYER" && id == "TDRC") {
					continue

				}

				columnIDs = append(columnIDs, id)

			}

		}

		for _, id := range columnIDs {
			data[0] = append(data[0], id)

			checkID := id
			if strings.Contains(id, "_") {
				idParts := strings.Split(id, "_")
				checkID = idParts[0]

			}

			value := raw[id]
			_, value2, valueType := GetTagValues(id, value)
			if valueType == "*tag.Comm" {
				data[1] = append(data[1], value2)

			} else if name, ok := newModTags.IsID(checkID); ok {
				data[1] = append(data[1], name)

			} else {
				data[1] = append(data[1], id)

			}

		}

	}
	raw["TFLT"] = fmt.Sprintf("%s", mp3.FileType())
	raw["Format"] = fmt.Sprintf("%s", mp3.Format())

	newRow := []string{}
	//for _, id := range columnIDs {
	for _, id := range data[0] {
		if value, ok := raw[id]; ok {
			value1, value2, _ := GetTagValues(id, value)
			if id == "PRIV" {
				newRow = append(newRow, value2)

			} else {
				newRow = append(newRow, value1)

			}

		} else {
			newRow = append(newRow, "")

		}

	}
	data[track+1] = newRow

	return nil

}

func GetTagValues(rawID string, rawValue any) (string, string, string) {
	if strings.Contains(rawID, "_") {
		rawIDParts := strings.Split(rawID, "_")
		rawID = rawIDParts[0]

	}

	var value1, value2 string
	valueType := fmt.Sprintf("%T", rawValue)
	if valueType == "*tag.Picture" {
		picture := rawValue.(*tag.Picture)
		if picture != nil {
			value1 = picture.MIMEType + " - " + picture.Type

		}

	} else if valueType == "*tag.Comm" {
		comm, ok := rawValue.(*tag.Comm)
		if !ok {
			slog.Error("rawValue.(*tag.Comm)|could not assert", "rawID", rawID, "rawValue", rawValue)
			return "", "", ""

		}

		value1 = comm.Text
		value2 = comm.Description

	} else if valueType == "[]uint8" {
		privs := rawValue.([]uint8)
		for i, b := range privs {
			if b == 0x00 {
				value2 = string(privs[:i])           // owner
				value1 = string([]byte(privs[i+1:])) // data

				break
			}

		}

	} else {
		value1 = fmt.Sprintf("%s", rawValue)

	}

	return value1, value2, valueType

}

func ReadTags(tracks []*Track, outFilenameOpt ...string) error {
	if len(outFilenameOpt) == 1 && outFilenameOpt[0] == "" {
		var emptyStringSlice []string
		outFilenameOpt = emptyStringSlice

	}

	var oFile *os.File = os.Stdout
	if len(outFilenameOpt) > 0 {
		var err error
		oFile, err = os.Create(outFilenameOpt[0])
		if err != nil {
			return err

		}

	}

	modTags := config.NewModTags()
	for _, track := range tracks {
		data := make(map[int][]string)
		readTag(data, track, modTags)
		stringy.PrintData(track.FilePath, data, oFile)

	}

	return nil

}

func readTag(data map[int][]string, track *Track, modTags *config.Tags) error {
	if track.mp3.V2 != nil {
		v2V1Tags := make(map[string]string)
		data[len(data)] = []string{"ID", "Name", "Value", "Type", "Extra", "V1.Name", "V1.Value"}
		for _, f := range track.mp3.V2.Frames {
			fID := f.ID()
			value := fmt.Sprintf("%v", f)
			if fID == "APIC" {
				value = fmt.Sprintf("%v", f.(*id3v2.PictureFrame).Description)

			}

			fType := reflections.AnyType(f)
			extraValue := ""
			if reflections.IsStruct(f) && fType != reflections.AnyType((*id3v2.TextFrame)(nil)) {
				if reflections.AnyType(f) == reflections.AnyType((*id3v2.CommentFrame)(nil)) {
					extraValue = fmt.Sprintf("CommentFrame.Body=%s", f.(*id3v2.CommentFrame).Text)

				} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.GenericFrame)(nil)) {
					extraValue = fmt.Sprintf("GenericFrame.Body=%s", f.(*id3v2.GenericFrame).Body)

				} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.PictureFrame)(nil)) {
					extraValue = fmt.Sprintf("PictureFrame.MIME=%s", f.(*id3v2.PictureFrame).MIME)

				} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.PrivFrame)(nil)) {
					extraValue = fmt.Sprintf("PrivFrame.Data=%s", f.(*id3v2.PrivFrame).Data)

				} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.UFIDFrame)(nil)) {
					extraValue = fmt.Sprintf("UFIDFrame.Identifier=%s", f.(*id3v2.UFIDFrame).Identifier)

				} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.UnsynchronisedLyricsFrame)(nil)) {
					extraValue = fmt.Sprintf("UnsynchronisedLyricsFrame.Text=%s", f.(*id3v2.UnsynchronisedLyricsFrame).Text)

				} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.URLFrame)(nil)) {
					extraValue = fmt.Sprintf("URLFrame.Text=%s", f.(*id3v2.UnsynchronisedLyricsFrame).Text)

				} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.UserTextFrame)(nil)) {
					extraValue = fmt.Sprintf("UserTextFrame.Value=%s", f.(*id3v2.UserTextFrame).Value)

				} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.UserURLFrame)(nil)) {
					extraValue = fmt.Sprintf("UserURLFrame.URL=%s", f.(*id3v2.UserURLFrame).URL)

				} else {
					extraValue = fmt.Sprintf("Frame=%#v", f)

				}

			}

			var v1Value string
			v1ID, ok := config.RequiredTags.Is2In1(fID)
			if ok && track.mp3.V1 != nil {
				if value, err := reflections.ReflectFieldByName(track.mp3.V1, v1ID); err == nil {
					v1Value = fmt.Sprintf("%v", value)
					v2V1Tags[fID] = v1ID

				}

			}

			data[len(data)] = []string{
				fID,
				modTags.Name(fID),
				value,
				fmt.Sprintf("%T", f),
				extraValue,
				v1ID,
				v1Value,
			}

		}

		for v2ID, v1ID := range config.RequiredTags.IdV2ToIDV1 {
			if _, ok := v2V1Tags[v2ID]; !ok {
				if v2ID == "TYER" {
					if _, ok := v2V1Tags["TDRC"]; ok {
						continue

					}

				}

				if value, err := reflections.ReflectFieldByName(track.mp3.V1, v1ID); err == nil {
					v1Value := fmt.Sprintf("%v", value)
					if v1ID == "Genre" {
						if genreNumber, err := strconv.Atoi(v1Value); err == nil {
							genreName := findV1GenreName(genreNumber)
							if genreName != "" {
								v1Value += " / " + genreName

							}

						}

					}

					if v1Value != "" {
						data[len(data)] = []string{
							v2ID,
							modTags.Name(v2ID),
							"",
							"",
							"",
							v1ID,
							v1Value,
						}

					}

				}

			}

		}

	}

	return nil

}

func ReadTag(filePath string) error {
	mp3, err := MP3(filePath)
	if err != nil {
		return err

	}

	if mp3.V2 != nil {
		modTags := config.NewModTags()
		for _, f := range mp3.V2.Frames {
			fID := f.ID()
			if fID != "APIC" {
				fmt.Printf("mp3.V2|ID=%v|Value=%v|Type=%T|", fID, f, f)
				fType := reflections.AnyType(f)
				if reflections.IsStruct(f) && fType != reflections.AnyType((*id3v2.TextFrame)(nil)) {
					if reflections.AnyType(f) == reflections.AnyType((*id3v2.GenericFrame)(nil)) {
						fmt.Printf("GenericFrame.Body=%s", f.(*id3v2.GenericFrame).Body)

					} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.PrivFrame)(nil)) {
						fmt.Printf("PrivFrame.Data=%s", f.(*id3v2.PrivFrame).Data)

					} else if reflections.AnyType(f) == reflections.AnyType((*id3v2.UFIDFrame)(nil)) {
						fmt.Printf("UFIDFrame.Data=%s", f.(*id3v2.UFIDFrame).Identifier)

					} else {
						fmt.Printf("Frame=%+#v", f)

					}

				} else {
					fmt.Printf("Name=%s", modTags.Name(fID))

				}
				fmt.Printf("\n")

			} else {
				fmt.Printf("mp3.V2|ID=%v|Value=|Type=%T|Name=%s\n", fID, f, modTags.Name("APIC"))

			}

		}

	}

	if mp3.V1 != nil {
		fmt.Printf("mp3.V1.Title=%v\n", mp3.V1.Title)
		fmt.Printf("mp3.V1.Artist=%v\n", mp3.V1.Artist)
		fmt.Printf("mp3.V1.Album=%v\n", mp3.V1.Album)
		fmt.Printf("mp3.V1.Comment=%v\n", mp3.V1.Comment)
		fmt.Printf("mp3.V1.Year=%v\n", mp3.V1.Year)
		fmt.Printf("mp3.V1.Track=%v\n", mp3.V1.Track)
		fmt.Printf("mp3.V1.Genre=%v\n", mp3.V1.Genre)

	}

	return nil

}

/*
	ID: TLAN, f: eng										Language
	ID: TRCK, f: 1/10										Track Number
	ID: TPE1, f: Courtney Marie Andrews						Artist
	ID: TIT2, f: Pendulum Swing								Title
	ID: TXXX, f: &{ISO-8859-1 Rip date 2026-04-30}			Rip Date
	ID: TYER, f: 2026										Year (TYER|TDRC replaces TYER)
	ID: TDAT, f: 0000										?? Month-Date
	ID: TXXX, f: &{ISO-8859-1 Source CDDA}					Source
	ID: TENC, f: TEAM ERP!									Encoded-by
	ID: TXXX, f: &{ISO-8859-1 Supplier TEAM ERP!}			Supplier
	ID: TSSE, f: Lame 3.100									Encoder Settings
	ID: TXXX, f: &{ISO-8859-1 Release type Album}			Release Type
	ID: TCON, f: Indie										Genre
	ID: TPUB, f: Loose Future Records						Publisher
	ID: TXXX, f: &{ISO-8859-1 Catalog # LFR0001CD}			Catalog #
	ID: TALB, f: Valentine									Album
	ID: PRIV, f: &{http://www.cdtag.com [2 0 244 30 91 0]}	http://www.cdtag.com
	ID: TPOS, f: 1/1										Disc Number
	ID: TPE2, f: Courtney Marie Andrews                     Album Artist
	ID: TOPE, f: Courtney Marie Andrews						Original Artist
	ID: COMM, f: &{ISO-8859-1 eng  Test Comment}            Comment
	ID: TCOM, f: My Composer                                Composer
	ID: APIC, f: &{ISO-8859-1 image/jpeg 3 00-courtney_marie_andrews-valentine-2026-erp.jpg
	MCDI
*/

func ReadTagType(filePath string) error {
	t, err := id3v2.ReadFile(filePath)
	if err != nil {
		return err

	}

	for i, f := range t.Frames {
		fmt.Printf("Frame[%d] %s (%T)\n", i, f.ID(), f)
		switch v := f.(type) {
		case *id3v2.TextFrame:
			fmt.Printf("  text: %q\n", v.String())
		case *id3v2.UserTextFrame:
			fmt.Printf("  TXXX desc=%q value=%q\n", v.Description, v.Value)
		case *id3v2.CommentFrame:
			fmt.Printf("  COMM lang=%q desc=%q text=%q\n", v.Language, v.Description, v.Text)
		case *id3v2.PictureFrame:
			fmt.Printf("  APIC mime=%q type=%d data=%d bytes\n", v.MIME, v.PictureType, len(v.Data))
		case *id3v2.URLFrame:
			fmt.Printf("  url: %q\n", v.URL)
		case *id3v2.PrivFrame:
			fmt.Printf("  PRIV owner=%q data=%d bytes\n", v.Owner, len(v.Data))
		case *id3v2.GenericFrame:
			// unknown / unsupported frames preserved verbatim
			fmt.Printf("  raw body=%d bytes\n", len(v.Body))
		}

	}

	return nil

}

func findV1Genre(genre string) string {
	index := 255
	if !taggerConfig.RemoveGenre {
		for i, v := range id3v1.Genres {
			if v == genre {
				index = i
				break

			}

		}

	}

	return strconv.Itoa(index)

}

func findV1GenreName(genreNumber int) string {
	genreName := ""
	if genreNumber >= 0 && genreNumber < len(id3v1.Genres) {
		genreName = id3v1.Genres[genreNumber]

	}

	return genreName

}

/*
TAG|filePath: /Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles/Courtney_Marie_Andrews-Valentine-2026-ERP/01-courtney_marie_andrews-pendulum_swing-erp.mp3
Frame[0] TLAN (*id3v2.TextFrame)
  text: "eng"
Frame[1] TRCK (*id3v2.TextFrame)
  text: "1/10"
Frame[2] TPE1 (*id3v2.TextFrame)
  text: "Courtney Marie Andrews"
Frame[3] TIT2 (*id3v2.TextFrame)
  text: "Pendulum Swing"
Frame[4] TXXX (*id3v2.UserTextFrame)
  TXXX desc="Rip date" value="2026-04-30"
Frame[5] TXXX (*id3v2.UserTextFrame)
  TXXX desc="Source" value="CDDA"
Frame[6] TENC (*id3v2.TextFrame)
  text: "TEAM ERP!"
Frame[7] TXXX (*id3v2.UserTextFrame)
  TXXX desc="Supplier" value="TEAM ERP!"
Frame[8] TSSE (*id3v2.TextFrame)
  text: "Lame 3.100"
Frame[9] TXXX (*id3v2.UserTextFrame)
  TXXX desc="Release type" value="Album"
Frame[10] TPUB (*id3v2.TextFrame)
  text: "Loose Future Records"
Frame[11] TXXX (*id3v2.UserTextFrame)
  TXXX desc="Catalog #" value="LFR0001CD"
Frame[12] TALB (*id3v2.TextFrame)
  text: "Valentine"
Frame[13] PRIV (*id3v2.PrivFrame)
  PRIV owner="http://www.cdtag.com" data=6 bytes
Frame[14] TPOS (*id3v2.TextFrame)
  text: "1/1"
Frame[15] TPE2 (*id3v2.TextFrame)
  text: "Courtney Marie Andrews"
Frame[16] TOPE (*id3v2.TextFrame)
  text: "Courtney Marie Andrews"
Frame[17] COMM (*id3v2.CommentFrame)
  COMM lang="eng" desc="" text="Test Comment"
Frame[18] TCOM (*id3v2.TextFrame)
  text: "My Composer"
Frame[19] APIC (*id3v2.PictureFrame)
  APIC mime="image/jpeg" type=3 data=234371 bytes
Frame[20] TYER (*id3v2.TextFrame)
  text: "2026"
Frame[21] TDAT (*id3v2.TextFrame)
  text: "0000"
Frame[22] TCON (*id3v2.TextFrame)
  text: "Indie"
*/
