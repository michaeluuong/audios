package mp3

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cabbagekobe/tunetag"
	"github.com/cabbagekobe/tunetag/id3v2"

	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/reflections"
	"github.com/michaeluuong/utilize/stringy"

	"github.com/michaeluuong/audios/audiostag/internal/a_global"
	"github.com/michaeluuong/audios/audiostag/internal/app/core/mp3/config"
	"github.com/michaeluuong/audios/audiostag/internal/pkg/aud_io"
)

// TODO need to remove (Bonus ), keep tags

const slogLevelNoOp = slog.LevelDebug - 4

// var audiosConfig = a_global.ConfigMan.ConfigPtr(new(a_global.AudiosConfig))
var audiosConfig, _ = a_global.FromConfigMan()
var taggerConfig = audiosConfig.MP3

// SetDefaultTags sets all tags to default values. Creates a cover file if one does not already exist (i.e. cover.jpg|folder.jpg|etc).
//   - dir is the parent directory of the files to tag
//   - filename is the file or regular expression representative of the files to tag ("" to default to ".*.mp3")
//   - tagAttr are tag attributes that can change tagging behavior
//   - -- tagAttr.IsPlaylist if true create a playlist with all files matching input parameters
//   - -- tagAttr.Audibles list of frames to change (e.g. artist=Jimi)
//
// Return an error if
//   - unable to set all of the tags to default values
func SetDefaultTags(dir, filename string, tagAttr *config.TagAttributes) error {
	slog.Log(context.TODO(), slogLevelNoOp+4, "started", "dir", dir, "filename", filename, "tagAttr", tagAttr)
	winman := aud_io.GetWinmanInstance()

	taggerConfig.TagAttr = tagAttr

	// Only want to do these one time
	slog.Info("AUD-2", "tagAttr.Audibles", tagAttr.Tags)
	idToTags := tagAttr.SplitTag()
	slog.Debug("SplitTag()", "idToTags", idToTags)
	replaceIDToValues, err := tagAttr.SplitReplacements()
	slog.Debug("SplitReplacements()", "replaceIDToValues", replaceIDToValues)
	if err != nil {
		slog.Error("tagAttr.SplitReplacements()|error splitting replacements", "tagAttr.Replacements", tagAttr.Replacements, "err", err)
		return err

	}
	_, err = tagAttr.SplitPreReplacements()

	addReplacementsToTagAttributes()
	if tagAttr.AlbumFolderName != "" {
		taggerConfig.AlbumFolderExp = tagAttr.AlbumFolderName

	}

	dirnames := filing.LsEntryName(dir, "-d", "add_dir")
	dirnames = append(dirnames, filepath.Join(dir, filename))
	slog.Info("directories to process", "dirnames", dirnames)
	for dirI, dir := range dirnames {
		fmt.Printf("dirI: %d, dir: %v\n", dirI, dir)
		artist, err := NewArtist()
		if err != nil {
			return err

		}
		artist.AddAll(dir)
		slog.Debug("POSE|After AddAll()")

		var playlist *Playlist
		if tagAttr.IsPlaylist {
			playlist = NewPlaylist()

		}

		for i, track := range artist.TrackIter() {
			winman.InfiniteProgressUpdate("Setting tags to default for " + track.Filename)
			//fmt.Printf("i: %d, track.Title(): %v\n", i, track.mp3.V2.Title())
			if dirI == 0 && i == 0 && !track.PicProcessed {
				winman.InfiniteProgressUpdate("Trying to find cover art")
				var coverSource string
				if audible, ok := idToTags["APIC"]; ok {
					coverSource = audible[0]

				}

				err := CreateCoverFile(track, coverSource)
				slog.Debug("POSE|After CreateCoverFile()", "track.Filename", track.Filename)
				if err != nil {
					slog.Error("CreateCoverFile()|could not find cover art", "filename", filename, "CoverSource", tagAttr.CoverSource, "err", err)

				}

				/*err = track.SetTag("APIC", coverSource, "no_replace")*/

			}

			if err := setDefaultTags(track); err != nil {
				slog.Error("setDefaultTags()|could not set tags to default", "OriginalFilePath", track.OriginalFilePath, "err", err)
				return err

			}

		}

		if !taggerConfig.TagAttr.NoDirectoryRename {
			artist.Rename()

		}

		stringy.PrintLine("-", 130, "\n")

		if tagAttr.IsPlaylist && playlist != nil {
			for _, disc := range artist.DiscIter() {
				if tagAttr.VariousArtists == "true" {
					playlist.VariousArtists = true

				}

				playlist.ScratchPlaylist(disc, tagAttr.PlaylistName)
				//playlist.ScratchPlaylist(disc.DestDir)
				playlist.ClearPlaylist()

				stringy.PrintLine("-", 130, "\n")

			}

		}

		if taggerConfig.ShowTags {
			for _, disc := range artist.DiscIter() {
				//ReadTags(disc.Tracks)

				//stringy.PrintLine("=", 130, "\n")
				PrintTags(disc.DestDir, "window", "false")

				stringy.PrintLine("-", 130, "\n")

			}

		}

	}

	winman.InfiniteProgressStop()

	return nil

}

func mapIDToValues(mp3 *tunetag.MP3) map[string]string {
	idToValues := make(map[string]string)

	for _, f := range mp3.V2.Frames {
		fID := f.ID()
		if fID == "APIC" {
			continue

		}

		fValue := fmt.Sprintf("%s", f)
		if fID == "COMM" {
			fValue = f.(*id3v2.CommentFrame).Text

		}

		idToValues[fID] = fValue

	}

	return idToValues

}

// setDefaultTags sets mp3 V1 and V2 tags to default values (i.e. IDs that are in requiredTags) and rename file.
//   - filePath the file with tags to set to default values
//   - tagAttr are tag attributes that can change tagging behavior
//
// NOTES
//   - always removes V1 comment
//   - remove total track number from track number (i.e. 1/10 becomes 1)
//   - Date/Year is always YYYY
//
// Return
//   - new filename
//   - new directory
//   - error
func setDefaultTags(track *Track) error {
	slog.Log(context.TODO(), slogLevelNoOp+4, "started", "track", track)
	_, filename := filepath.Split(track.FilePath)

	mp3 := track.mp3
	idToFrames := mapIDToFrames(mp3)
	idToValues := mapIDToValues(mp3)

	// Add a cover art image if there isn't already an image in the tag
	if err := AddAPIC(mp3, track.OriginalFilePath, false); err != nil {
		slog.Error("AddAPIC()|could not add cover art", "filePath", track.FilePath, "err", err)
		return err

	}

	// Change the case of Title, Artist and Album
	if isCased := setCase(mp3); !isCased {
		slog.Debug("setCase()|no case set", "filename", filename, "StringCase", fmt.Sprintf("%s", taggerConfig.TagAttr.StringCase))

	}

	idToTags := taggerConfig.TagAttr.SplitTag()

	processArtist(track, taggerConfig.TagAttr.VariousArtists, idToTags)
	processReplacements(track, idToValues)

	if mp3.V1.Comment != "" { // Always clear Id3V1 Comment field
		mp3.V1.Comment = ""

	}

	for fID, f := range idToFrames {
		if fID == "APIC" {
			continue

		}

		var fValue string = fmt.Sprintf("%s", f)
		if fID == "COMM" {
			if _, ok := idToTags[fID]; !ok {
				mp3.V2.RemoveFrames("COMM")

			}

		} else if !config.RequiredTags.IsID(fID) && !taggerConfig.TagAttr.IsKeeper(fID) {
			mp3.V2.RemoveFrames(fID)

		} else {
			newFValue := "default"
			switch fID {
			case "TCON": // Genre
				if taggerConfig.RemoveGenre {
					newFValue = ""
					mp3.V1.Genre = 255

				}

			case "TDRC": // Date
				if len(fValue) > 4 {
					newFValue = fValue[:4]

				}

			}

			if newFValue != "default" {
				if fValue != "" {
					mp3.V2.SetText(fID, newFValue)

				} else {
					mp3.V2.RemoveFrames(fID)

				}

			}

		}

	}

	processTags(track, idToTags)

	if err := track.Write(); err != nil {
		slog.Error("track.Write()|could not write tags", "err", err)
		return err

	}

	err := track.Rename()

	//slog.Log(context.TODO(), slogLevelNoOp+4, "finished", "newFilename", newFilename, "newChildFolder", newChildFolder)
	return err

}

func mapIDToFrames(mp3 *tunetag.MP3) map[string]id3v2.Frame {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "mp3", mp3)
	idToFrames := make(map[string]id3v2.Frame)

	for _, f := range mp3.V2.Frames {
		if f.ID() != "APIC" {
			idToFrames[f.ID()] = f

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished", "idToFrames", idToFrames)
	return idToFrames

}

func mapFrameNamesToValues(filePath string) map[string]string {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "filePath", filePath)
	namesToValues := make(map[string]string)

	mp3, err := MP3(filePath)
	if err != nil {
		return namesToValues

	}

	modTags := config.NewModTags()
	for _, f := range mp3.V2.Frames {
		fValue := fmt.Sprintf("%s", f)
		if name, ok := modTags.IsID(f.ID()); ok && f.ID() != "APIC" {
			namesToValues[name] = fValue

		}

	}

	// Special
	if val, ok := namesToValues["Year"]; !ok || val == "" {
		if val2, ok2 := namesToValues["Recording Time"]; ok2 && val2 != "" {
			namesToValues["Year"] = namesToValues["Recording Time"]

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished", "namesToValues", namesToValues)
	return namesToValues

}

func SetTagsDir(dir string, tagAttrOpt ...*config.TagAttributes) error {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "dir", dir)
	if len(tagAttrOpt) > 0 && tagAttrOpt[0] != nil {
		taggerConfig.TagAttr = tagAttrOpt[0]

	}

	dirnames := filing.LsEntryName(dir, ".*\\.mp3", "add_dir")

	var idToTags map[string][]string = taggerConfig.TagAttr.SplitTag()

	for _, filePath := range dirnames {
		track, err := NewTrack(filePath)
		if err != nil {
			slog.Error("NewTrack()|could not open mp3", "filePath", filePath, "err", err)
			return err
		}

		if err := SetTagsFile(track, idToTags); err != nil {
			slog.Error("SetTagsFile()|could not set tags", "idToTags", idToTags)
			return err

		}

		if err := track.mp3.V1.WriteFile(filePath); err != nil {
			slog.Error("mp3.V1.WriteFile()|could not write tags", "filePath", filePath)
			return err

		}

		if err := track.mp3.V2.WriteFile(filePath); err != nil {
			slog.Error("mp3.V2.WriteFile()|could not write tags", "filePath", filePath)
			return err

		}

	}

	PrintTags(dir)

	slog.Log(context.TODO(), slogLevelNoOp, "finished")
	return nil

}

func SetTagsFile(track *Track, idToTags map[string][]string) error {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "track", track, "idToTags", idToTags)
	for id, values := range idToTags {
		var tagValue, tagValue2 string = values[0], ""
		if len(values) >= 3 {
			tagValue2 = values[1]

		}

		err := track.SetTag(id, tagValue, tagValue2)
		if err != nil {
			slog.Error("SetTag()|could not set tag", "id", id, "tagValue", tagValue, "tagValue2", tagValue2)
			return err

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")
	return nil

}

func SetTag(track *Track, id, tagValue string, tagValuesOpt ...string) error {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "track", track, "id", id, "tagValue", tagValue, "tagValuesOpt", tagValuesOpt)
	tagValue2 := ""
	if len(tagValuesOpt) > 0 {
		tagValue2 = tagValuesOpt[0]

	}

	mp3 := track.mp3
	altTagType := config.NewAltTagType()
	if tagType, ok := altTagType.IDHasType(id); ok {
		if tagType == reflections.AnyType((*id3v2.PictureFrame)(nil)) { // APIC
			if err := CreateCoverFile(track, tagValue); err != nil {
				return err

			}

			err := AddAPIC(mp3, track.OriginalFilePath, true)
			if err != nil {
				slog.Error("AddAPIC()|could not add picture frame", "err", err)
				return err

			}

		} else if tagType == reflections.AnyType((*id3v2.CommentFrame)(nil)) { // COMM
			addCommentFrame(mp3, tagValue)

		} else if tagType == reflections.AnyType((*id3v2.GenericFrame)(nil)) { // IPLS|GRP1|MVIN|MVNM|OWNE|PCST|POPM|SYLT
			addGenericFrame(mp3, id, tagValue)

		} else if tagType == reflections.AnyType((*id3v2.UserTextFrame)(nil)) { // TXXX
			addUserTextFrame(mp3, tagValue, tagValue2)

		} else if tagType == reflections.AnyType((*id3v2.URLFrame)(nil)) { // WCOP|WFED|WOAF|WOAR|WORS|WPAY|WPUB
			addURLFrame(mp3, id, tagValue)

		} else if tagType == reflections.AnyType((*id3v2.PrivFrame)(nil)) { // PRIV
			addPrivFrame(mp3, id, tagValue, tagValue2)

		} else if tagType == reflections.AnyType((*id3v2.UFIDFrame)(nil)) { // UFID
			addUFIDFrame(mp3, id, tagValue, tagValue2)

		} else if tagType == reflections.AnyType((*id3v2.UnsynchronisedLyricsFrame)(nil)) { // USLT
			addUnsynchronisedLyricsFrame(mp3, id, tagValue, tagValue2)

		} else if tagType == reflections.AnyType((*id3v2.UserURLFrame)(nil)) { // WXXX
			addUserURLFrame(mp3, id, tagValue, tagValue2)

		}

	} else {
		mp3.V2.SetText(id, tagValue)

		if config.RequiredTags.IsCopyTo(id) { // Copy this value to other fields
			for _, field := range config.RequiredTags.CopyToFields[id] {
				mp3.V2.SetText(field, tagValue)

			}
		}

		f := &id3v2.TextFrame{
			FrameID:  id,
			Encoding: id3v2.EncUTF8,
			Text:     []string{tagValue},
		}
		copyToID1(mp3, f)

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")
	return nil

}

// copyToID1 copies an mp3.V2 tag to mp3.V1 if applicable (i.e. there is a corresponding tag specified in requiredTags)
//   - mp3 is the MP3 tag struct containing the tags
//   - f is the frame to copy from V2 to V1
func copyToID1(mp3 *tunetag.MP3, f id3v2.Frame) {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "mp3", mp3, "f", f)
	fID := f.ID()
	//regexCurly := regexp.MustCompile("’")
	if _, ok := config.RequiredTags.Is2In1(fID); ok {
		f1Value := ""
		var fValue string = fmt.Sprintf("%s", f)
		if fID == "COMM" && fValue != "" { // Comment
			f1Value = mp3.V2.Comment()

		} else if fID == "TCON" {
			if fValue == "" {
				f1Value = strconv.Itoa(255)

			} else {
				f1Value = findV1Genre(fValue)

			}

		} else {
			f1Value = fValue

		}
		//f1Value = regexCurly.ReplaceAllString(f1Value, "'")

		v1Field := config.RequiredTags.IdV2ToIDV1[fID]
		slog.Debug("copying mp3.V2 tag to mp3.V1", "fID", fID, "fValue", fValue, "v1Field", v1Field)
		reflections.SetStructField(mp3.V1, v1Field, f1Value)

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")

}

// CreateCoverFile creates a cover file if one does not already exist.
//   - mp3FilePath is the path to the mp3 files where the cover file will be created
//
// Return an error if unable to create the cover art file.
// func CreateCoverFile(mp3FilePath string, tagAttr *TagAttributes) error {
// func CreateCoverFile(mp3FilePath, coverSource string, tagAttrOpt ...*TagAttributes) error {
func CreateCoverFile(track *Track, coverSource string) error {
	slog.Log(context.TODO(), slogLevelNoOp+4, "started", "track", track, "coverSource", coverSource)
	if coverSource == "" && taggerConfig.TagAttr != nil && taggerConfig.TagAttr.CoverSource != "" {
		coverSource = taggerConfig.TagAttr.CoverSource

	}

	slog.Debug("CreateCoverFile()|coverSource", "coverSource", coverSource)

	var coverFile string
	var err error
	if !CoverExists(track.OriginalDir) && len(track.mp3.V2.PictureFrames()) == 0 {
		coverFilenameNoExt := taggerConfig.CoverFilenameNoExt
		if coverSource != "" {
			if coverFilenameNoExt == "" {
				coverFilenameNoExt = "cover"

			}

			if matched, _ := regexp.MatchString("^https?://", coverSource); matched { // URL
				slog.Info("GETting image from URL", "coverSource", coverSource)
				coverFilenameNoExt += "69"
				coverFile, err = DownloadPicture(coverSource, track.OriginalDir, coverFilenameNoExt)
				if err != nil {
					slog.Error("DownloadPicture()|could not GET cover art", "coverFile", coverFile, "err", err)
					return err

				}

			} else if filing.Exists(coverSource) { // Local file
				slog.Info("Image in local file", "coverSource", coverSource)
				ext := filepath.Ext(coverSource)
				coverFilenameNoExt += "68"
				coverFile = filepath.Join(track.OriginalDir, coverFilenameNoExt+ext)
				fmt.Printf("CreateCoverFile()|coverFile: %s\n", coverFile)
				err = filing.CopyFile(coverSource, coverFile)
				if err != nil {
					slog.Error("filing.CopyFile()|could not copy file", "coverSource", coverSource, "coverFile", coverFile)
					return err

				}

			}

		}

		if coverFile == "" {
			mp3 := track.mp3
			originalAlbum := track.Value("TOAL")
			artist, album := mp3.V2.Artist(), mp3.V2.Album()
			if originalAlbum != "" {
				album = originalAlbum

			}
			if taggerConfig.TagAttr != nil {
				if taggerConfig.TagAttr.CoverArtist != "" {
					artist = taggerConfig.TagAttr.CoverArtist

				}

				if taggerConfig.TagAttr.CoverAlbum != "" {
					album = taggerConfig.TagAttr.CoverAlbum

				}

			}

			if _, err := CoverArt(artist, album, track.OriginalDir, coverFilenameNoExt); err != nil {
				slog.Error("CoverArt()|could not GET cover art", "artist", artist, "album", album, "originalDir", track.OriginalDir, "err", err)
				return err

			}

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")
	return nil

}

func addReplacementsToTagAttributes() {
	slog.Log(context.TODO(), slogLevelNoOp, "started")
	if taggerConfig.TagReplacements != "" {
		taggerConfig.TagAttr.SplitReplacements(taggerConfig.TagReplacements)

	}

	if taggerConfig.FeaturingFix != "" {
		featuringFix := "( [(\\[]?)([fF]eaturing|[fF]eat\\.?|[fF]t\\.?) "
		taggerConfig.TagAttr.AddReplacement("TIT2", featuringFix, "${1}"+taggerConfig.FeaturingFix+" ")
		slog.Debug("AddReplacement()|adding featuring replacement for Title", "featuring_fix", taggerConfig.FeaturingFix, "featuringFix", featuringFix)
		if taggerConfig.FeaturingParen {
			parenFeaturingFix := " (\\]?" + taggerConfig.FeaturingFix + " [^\\[]+)([()]?|$)"
			slog.Debug("AddReplacement()|adding parentheses around featuring for Title", "parenFeaturingFix", parenFeaturingFix)
			taggerConfig.TagAttr.AddReplacement("TIT2", parenFeaturingFix, " ($1)$2")

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")

}

func setCase(mp3 *tunetag.MP3) bool {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "mp3", mp3)
	isCased := false
	if taggerConfig.TagAttr.StringCase != stringy.DefaultCase {
		mp3.V2.SetTitle(stringy.CaseString(mp3.V2.Title(), taggerConfig.TagAttr.StringCase))
		mp3.V2.SetArtist(stringy.CaseString(mp3.V2.Artist(), taggerConfig.TagAttr.StringCase))
		mp3.V2.SetAlbum(stringy.CaseString(mp3.V2.Album(), taggerConfig.TagAttr.StringCase))

		isCased = true

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished", "isCased", isCased)
	return isCased

}

// processReplacements performs replacements on tags by regular expression.
//
// NOTE: replacements cannot be performed on tags that aren't text based like the cover art (APIC) frame.
// func processReplacements(mp3 *tunetag.MP3, idToFrames map[string]id3v2.Frame, tagAttr *TagAttributes) error {
func processReplacements(track *Track, idToValues map[string]string) error {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "track", track, "idToValues", idToValues)
	idToReplacements, err := taggerConfig.TagAttr.SplitReplacements()
	if err != nil {
		slog.Error("SplitReplacements()|could not process Replacements", "err", err)
		return err

	}

	for id, replacements := range idToReplacements {
		if id == "APIC" {
			slog.Error("input|cannot perform a replacement on the APIC (cover art) frame")
			//return errors.New("cannot perform  a replacment on the APIC (cover art) frame")
			continue

		} else if id == "TALB" {
			slog.Error("input|cannot perform a replacement on the TALB (Album) frame")
			//return errors.New("cannot perform  a replacment on the APIC (cover art) frame")
			continue

		}

		if fValue, ok := track.idToValue[id]; ok {
			for _, replacement := range replacements {
				replaceRe, replaceExp := replacement.RegEx, replacement.Replace
				replaceRegexp, err := regexp.Compile(replaceRe)
				if err != nil {
					slog.Error("regexp.Compile()|problem with regular expression", "id", id, "replaceRe", replaceRe, "replaceExp", replaceExp)
					return err

				}
				fValue = replaceRegexp.ReplaceAllString(fValue, replaceExp)

				slog.Debug("ReplaceAllString()|replaced", "id", id, "fValue", fValue, "replaceRe", replaceRe, "replaceExp", replaceExp)
				track.SetTag(id, fValue)
				idToValues[id] = fValue

			}

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")
	return nil

}

// processArtist
//
// NOTE: if this is a variousArtists album this will remove extra artists
func processArtist(track *Track, variousArtists string, idToTags map[string][]string) {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "track", track, "variousArtists", variousArtists, "idToTags", idToTags)
	mp3 := track.mp3
	slog.Debug("POES", "idToTags", idToTags)
	if variousArtists == "true" { // Append artist to title field
		artist := strings.ReplaceAll(mp3.V2.Artist(), "-", "+")
		newTitle := strings.ReplaceAll(mp3.V2.Title(), "-", "+") + " - " + artist
		track.SetTag("TIT2", newTitle)
		track.SetTag("TPE1", track.mp3.V2.Album())
		track.SetTag("TOPE", artist)

	} else if _, ok := idToTags["TPE1"]; !ok { // Add artist so processTags can remove extra artists
		artist := mp3.V2.Artist()
		xEx := " [xX] "
		xRe := regexp.MustCompile(xEx)
		slog.Debug("POES", "artist", artist, "xEx", xEx)
		if xRe.MatchString(artist) {
			track.SetTag("TIT2", track.mp3.V2.Title()+" ("+artist+")")
			oneX := xRe.Split(artist, -1)
			slog.Debug("POES", "title", track.mp3.V2.Title(), "oneX", oneX)
			track.SetTag("TPE1", oneX[0])

		} else if taggerConfig.TagAttr.SingleArtist != "true" {
			artistSplit := strings.Split(artist, ",")
			artists := []string{artistSplit[0]}
			if len(artistSplit) > 1 {
				idToTags["TPE1"] = artists

			}

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")

}

func processTags(track *Track, idToTags map[string][]string) {
	for id, values := range idToTags {
		// Skip Album changes as it needs to be done in the beginning before directories are set up
		if id == "TALB" {
			slog.Debug("input|album changes need to be done in the beginning before directories are set up", "id", id)
			continue

		}

		newValue := values[0]

		var value2 string
		if len(values) > 1 {
			value2 = values[1]

		}

		if taggerConfig.TagAttr.StringCase != stringy.DefaultCase {
			newValue = stringy.CaseString(values[0], taggerConfig.TagAttr.StringCase)

		}

		slog.Debug("POES", "SingleArtist", taggerConfig.TagAttr.SingleArtist)
		if id == "TPE1" && taggerConfig.TagAttr.VariousArtists != "true" && strings.Contains(track.mp3.V2.Artist(), ",") &&
			taggerConfig.TagAttr.SingleArtist != "true" {
			titleWith := artistToTitleWith(track.mp3, newValue)
			track.SetTag("TIT2", titleWith)

		}

		slog.Debug("SetTag()|setting tag", "id", id, "originalValue", track.idToValue[id], "newValue", newValue, "value2", value2)
		track.SetTag(id, newValue, value2)

	}

}

func artistToTitleWith(mp3 *tunetag.MP3, newArtist string) string {
	artist := mp3.V2.Artist()
	artistToTitle := taggerConfig.ArtistToTitle
	if artistToTitle == "" {
		artistToTitle = "with"

	}

	withString := stringy.ItemSentence(artist, ",", " ("+artistToTitle+" ", ")", map[string]bool{newArtist: true})

	return strings.TrimSpace(mp3.V2.Title()) + withString

}
