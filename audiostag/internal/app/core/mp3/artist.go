package mp3

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/cabbagekobe/tunetag"
	"github.com/cabbagekobe/tunetag/id3v1"
	"github.com/cabbagekobe/tunetag/id3v2"

	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/reflections"
	"github.com/michaeluuong/utilize/stringy"

	"github.com/michaeluuong/audios/audiostag/internal/app/core/mp3/config"
)

// Artist -> Album -> Disc -> Track

const (
	slogLevelArtist = slogLevelNoOp
)

type DirLevel int

const (
	DirLevelArtist DirLevel = iota
	DirLevelAlbum
	DirLevelDisc
	DirLevelTrack
)

var altTagType = config.NewAltTagType()

// NewArtist creats an Artist object with all maps initialized.
//   - artistOpt optional Artist object
//
// Return a pointer to an Artist object with all map fields initialized or an error if the map(s) could not be initialized
func NewArtist(artistOpt ...*Artist) (*Artist, error) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "artistOpt", artistOpt)
	artist := &Artist{}
	if len(artistOpt) > 0 {
		artist = artistOpt[0]

	}

	if err := reflections.InitializeStruct(artist); err != nil {
		return nil, err

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished", "artist", artist)
	return artist, nil

}

type Artist struct {
	DestDir                     string `json:"dest_dir"`
	Name                        string `json:"name"`
	OriginalDir                 string `json:"OriginalDir"`
	albums                      []*Album
	nameToAlbum                 map[string]*Album
	albumNamesToAlbumNum        map[string]int
	originalFilePathToAlbumName map[string]string
	baseDestDir                 string
	lastDestDir                 string
}

func (a *Artist) String() string {
	slog.Log(context.TODO(), slogLevelArtist, "started")

	return reflections.StructString(a, "newline", "\t")

}

func (a *Artist) PrintDestDirs() {
	fmt.Printf("\n\na.DestDir: %s,\na.baseDestDir: %s,\na.lastDestDir: %s\n", a.DestDir, a.lastDestDir, a.baseDestDir)
	for _, album := range a.albums {
		fmt.Printf("\talbum.DestDir: %s,\n\talbum.baseDestDir: %s,\n\talbum.lastDestDir: %s\n", album.DestDir, album.lastDestDir, album.baseDestDir)
		for _, disc := range album.numToDisc {
			fmt.Printf("\t\tdisc.DestDir: %s,\n\t\tdisc.lastDestDir: %s\n", disc.DestDir, disc.lastDestDir)
			for _, track := range disc.Tracks {
				fmt.Printf("\t\t\ttrack: %s\n", track.DestDir)

			}

		}

	}

}

func (a *Artist) AddAll(dir string, tagAttrOpt ...*config.TagAttributes) error {
	slog.Log(context.TODO(), slogLevelArtist, "started", "dir", dir)
	if len(tagAttrOpt) > 0 && tagAttrOpt[0] != nil {
		taggerConfig.TagAttr = tagAttrOpt[0]

	}

	filename := ".*\\.mp3"

	maxFileLimit := taggerConfig.MaxFileLimit
	if maxFileLimit == 0 {
		maxFileLimit = 100

	}

	filePaths := filing.LsEntryName(dir, filename, "add_dir")
	filenamesLen := len(filePaths)
	if filenamesLen == 0 {
		slog.Error("filing.LsEntryName()|nothing to process", "dir", dir, "filename", filename)
		return fmt.Errorf("nothing to process, dir: %s, filename: %s", dir, filename)

	} else if filenamesLen > maxFileLimit {
		slog.Error("filing.LsEntryName()|too many files to process", "dir", dir, "filename", filename, "filenamesLen", filenamesLen)
		return fmt.Errorf("too many files to process, dir:, %s, filename: %s, filenamesLen: %d", dir, filename, filenamesLen)

	}

	for i, filePath := range filePaths {
		//stringy.PrintLine("=", 80)
		//fmt.Printf("%s\n", filePath)
		a.Add(filePath)
		if i == 1 {
			//os.Exit(1)
		}

	}

	if taggerConfig.TotalTracks {
		a.SetTotalTrackNumbers(true)

	}

	//fmt.Printf("Artist.AddAll()|a: %s\n", a)
	/*fmt.Printf("Artist.AddAll()|a.albums: %s\n", a.albums)
	for _, album := range a.albums {
		fmt.Printf("Artist.AddAll|\talbum.numToDisc: %s\n", album.numToDisc)

	}*/
	//a.PrintDestDirs()

	if taggerConfig.TagAttr.MultiDisc == "" && a.TotalDiscs() > 1 {
		taggerConfig.TagAttr.MultiDisc = "true"

	}

	const eightyPercent float64 = 0.80
	var eightyPercentOfTotal float64 = math.Ceil(float64(a.TotalTracks()) * eightyPercent)
	if taggerConfig.TagAttr.VariousArtists == "" && a.findCommonArtist() == "" && a.TotalArtists() > 1 && a.TotalArtists() >= int(eightyPercentOfTotal) {
		taggerConfig.TagAttr.VariousArtists = "true"

	}

	slog.Debug("POES", "TotalArtists()", a.TotalArtists(), "TotalTracks()", a.TotalTracks(), "findCommonArtist()", a.findCommonArtist(), "a.Name", a.Name)
	//if a.TotalArtists() == 1 && !strings.Contains(a.Name, ",") && !taggerConfig.TagAttr.NoArtistSplit {
	if a.TotalArtists() == 1 && taggerConfig.TagAttr.SingleArtist == "" {
		taggerConfig.TagAttr.SingleArtist = "true"

	}
	slog.Debug("POES", "VariousArtists", taggerConfig.TagAttr.VariousArtists, "SingleArtist", taggerConfig.TagAttr.SingleArtist)

	slog.Log(context.TODO(), slogLevelArtist, "finished")
	return nil

}

func (a *Artist) findCommonArtist() string {
	artistNum := make(map[string]int)
	fullArtistNum := make(map[string]int)
	var artist string
	for _, track := range a.TrackIter() {
		curArtist := track.mp3.V2.Artist()
		if artist == "" {
			artist = curArtist

		}

		fullArtistNum[curArtist]++
		for artistPart := range strings.SplitSeq(curArtist, ",") {
			artistNum[strings.TrimSpace(artistPart)]++

		}

	}
	slog.Debug("POES", "artistNum", artistNum)

	if len(fullArtistNum) == 1 {
		return artist

	} else {
		for art, num := range artistNum {
			if num == a.TotalTracks() {
				return art

			}

		}

	}

	return ""

}

func (a *Artist) Add(filePath string) (*Track, error) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "filePath", filePath)

	track, err := NewTrack(filePath)
	if err != nil {
		slog.Error("NewTrack()|could not create new track", "filePath", filePath, "err", err)
		return nil, err

	}

	if taggerConfig.TagAttr != nil {
		track.processPreConditions()
		track.processReplacements(true)

	}

	a.setName(track.mp3.V2.Artist())
	a.addToAlbum(track)

	// Set directories last
	a.setDestDirs(track)

	slog.Log(context.TODO(), slogLevelArtist, "finished", "track", track)
	return track, nil

}

func (a *Artist) addToAlbum(track *Track) (*Album, error) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "track", track)

	albumName := track.mp3.V2.Album()

	album := a.Album(albumName)
	//fmt.Printf("addToAlbum()|album: %s, albumName: %s\n", album, albumName)
	if album == nil {
		var err error
		album, err = NewAlbum(albumName, track.DestDir)
		if err != nil {
			return nil, err

		}

		albumsLen := len(a.albums)
		a.albums = append(a.albums, album)
		a.albumNamesToAlbumNum[albumName] = albumsLen

	}

	a.originalFilePathToAlbumName[track.FilePath] = albumName

	if _, err := album.addToDisc2(track); err != nil {
		return nil, err

	}

	album.addArtistName(track.mp3.V2.Artist())

	slog.Log(context.TODO(), slogLevelArtist, "finished", "album", album)
	return album, nil

}

func (a *Artist) Album(name string) *Album {
	slog.Log(context.TODO(), slogLevelArtist, "started", "name", name)

	var album *Album

	if albumNum, ok := a.IsAlbum(name); ok {
		album = a.albums[albumNum]

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished", "album", album)
	return album

}

func (a *Artist) IsAlbum(albumName string) (int, bool) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "albumName", albumName)

	albumNum, ok := a.albumNamesToAlbumNum[albumName]
	if !ok {
		albumNum = -1

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished", "albumNum", albumNum, "ok", ok)

	return albumNum, ok

}

func (a *Artist) SetDestDir(filePath string) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "filePath", filePath)

	artistDir := filepath.Dir(filepath.Clean(filePath))
	artistBase := filepath.Base(filepath.Clean(artistDir))
	a.DestDir = artistDir
	a.baseDestDir = artistBase

	slog.Log(context.TODO(), slogLevelArtist, "finished", "DestDir", a.DestDir)

}

func (a *Artist) setDestDirs(track *Track, dirLevelOpt ...DirLevel) {
	if len(dirLevelOpt) > 0 && dirLevelOpt[0] == DirLevelArtist {
		a.setDestDir(track, true)

	} else {
		a.setDestDir(track, false)

	}

	var album *Album

	albumName := track.Album()
	album = a.Album(albumName)
	album.setDestDir(track)

	discNum := track.DiscNumber()
	disc := album.Disc(discNum)
	disc.setDestDir(track)

}

func (a *Artist) setDestDir(track *Track, isRename bool) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "track", track)

	filePath := track.FilePath
	albumDir := filepath.Dir(filepath.Clean(filePath))
	artistDir := filepath.Dir(filepath.Clean(albumDir))
	artistBase := track.mp3.V2.Artist()
	a.lastDestDir = a.DestDir
	if taggerConfig.ArtistFolderExp != "" && isRename && !taggerConfig.TagAttr.NoDirectoryRename {
		parentDir := filepath.Dir(filepath.Clean(artistDir))
		if track.MultiDiscAlbumName() != "" {
			parentDir = filepath.Dir(filepath.Clean(parentDir))

		}

		artistBase = stringy.ReplaceUserString(taggerConfig.ArtistFolderExp, track.nameToValue, stringy.TitleCase)
		a.DestDir, _ = filing.NextDir(filepath.Join(parentDir, artistBase))

	} else if a.DestDir == "" || a.DestDir != track.OriginalDir {
		a.DestDir = track.OriginalDir

	}

	if a.OriginalDir == "" {
		a.OriginalDir = a.DestDir

	}

	a.baseDestDir = artistBase

	trackDir := filepath.Join(a.DestDir, track.Filename)
	track.SetDestDir(trackDir)

	slog.Log(context.TODO(), slogLevelArtist, "finished", "DestDir", a.DestDir)

}

func (a *Artist) setName(artistName string) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "artistName", artistName)

	if a.Name != artistName {
		a.Name = artistName

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished", "a.Name", a.Name)

}

func (a *Artist) Albums() []*Album {
	slog.Log(context.TODO(), slogLevelArtist, "started")

	slog.Log(context.TODO(), slogLevelArtist, "finished") //, "albums", a.albums)
	return a.albums

}

func (a *Artist) IsAlbumF(filePath string) (int, bool) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "filePath", filePath)

	var albumNum int
	if albumName, ok := a.originalFilePathToAlbumName[filePath]; ok {
		albumNum, ok = a.albumNamesToAlbumNum[albumName]

		return albumNum, ok

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished", "albumNum", -1, "ok", false)
	return -1, false

}

func (a *Artist) AlbumF(filePath string) *Album {
	slog.Log(context.TODO(), slogLevelArtist, "started", "filePath", filePath)

	var album *Album
	if albumNum, ok := a.IsAlbumF(filePath); ok {
		album = a.albums[albumNum]

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished", "album", album)
	return album

}

func (a *Artist) AddAlbum(filePath string) (*Album, error) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "filePath", filePath)

	track, err := NewTrack(filePath)
	if err != nil {
		slog.Error("NewTrack()|could not create track", "filePath", filePath, "error", err)
		return nil, err

	}

	album, err := a.addToAlbum(track)
	if err != nil {
		return nil, err

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished", "album", album)
	return album, nil

}

func (a *Artist) Track(filePath string) (*Track, error) {
	album := a.AlbumF(filePath)
	disc := album.DiscF(filePath)
	track, err := disc.TrackF(filePath)

	return track, err

}

func (a *Artist) TotalTracks() int {
	totalTracks := 0
	for _, album := range a.albums {
		totalTracks += album.TotalTracks()

	}

	return totalTracks

}

func (a *Artist) TotalArtists() int {
	totalArtists := 0
	for _, album := range a.albums {
		totalArtists += len(album.artistNames)

	}

	return totalArtists

}

func (a *Artist) TotalDiscs() int {
	totalDiscs := 0
	for _, album := range a.albums {
		totalDiscs += len(album.discs)

	}

	return totalDiscs

}

func (a *Artist) SetTotalTrackNumbers(setTotal bool) {
	for _, album := range a.albums {
		album.SetTotalTrackNumbers(setTotal)

	}

}

func (a *Artist) Rename() error {
	var trackDir string
	fmt.Printf("Rename()|a.DestDir: %s\n", a.DestDir)
	for _, track := range a.TrackIter() {
		if trackDir == "" {
			if taggerConfig.ArtworkRegex() != "" {
				filenames := filing.LsEntryName(a.OriginalDir, taggerConfig.ArtworkRegex())

				for _, filename := range filenames {
					sourcePath := filepath.Join(a.OriginalDir, filename)
					albumDir := a.albums[0].DestDir
					destPath := filepath.Join(albumDir, filename)
					os.Rename(sourcePath, destPath)

				}

			}

		}
		a.setDestDirs(track, DirLevelArtist)

	}

	fmt.Printf("Rename()|a.DestDir: %s\n", a.DestDir)
	err := os.Rename(a.OriginalDir, a.DestDir)
	if err != nil {
		return err

	}

	return nil

}

func (a *Artist) DiscIter() iter.Seq2[int, *Disc] {
	return func(yield func(int, *Disc) bool) {
		for _, album := range a.albums {
			discI := 0
			discNums := slices.Collect(maps.Keys(album.numToDisc))
			slices.Sort(discNums)
			for _, discNum := range discNums {
				disc := album.numToDisc[discNum]
				discI++
				if !yield(discI, disc) {
					return

				}

			}

		}

	}

}

func (a *Artist) TrackIter() iter.Seq2[int, *Track] {
	return func(yield func(int, *Track) bool) {
		for _, album := range a.albums {
			discNums := slices.Collect(maps.Keys(album.numToDisc))
			slices.Sort(discNums)

			trackI := 0
			for _, discNum := range discNums {
				disc := album.numToDisc[discNum]
				for _, track := range disc.Tracks {
					if !yield(trackI, track) {
						return

					}

					trackI++

				}

			}

		}

	}

}

func (a *Artist) TrackIter2() iter.Seq2[int, *Track] {
	return func(yield func(int, *Track) bool) {
		for _, album := range a.albums {
			trackI := 0
			for _, disc := range album.discs {
				for _, track := range disc.Tracks {
					if !yield(trackI, track) {
						return

					}

					trackI++

				}

			}

		}

	}

}

// --------------------------------------------------------------------------------------------------
func NewAlbum(albumName, albumDestDir string) (*Album, error) {
	album := &Album{
		Name:    albumName,
		DestDir: albumDestDir,
		//artistNames: make(map[string]bool),
		//discs:             []*Disc{},
		//filePathToDiscNum: make(map[string]int),
	}

	if err := reflections.InitializeStruct(album); err != nil {
		return nil, err

	}

	return album, nil

}

type Album struct {
	Name              string `json:"name"`
	Cover             string `json:"cover"`
	DestDir           string `json:"dest_dir"`
	artistNames       map[string]bool
	discs             []*Disc
	numToDisc         map[int]*Disc
	filePathToDiscNum map[string]int
	baseDestDir       string
	lastDestDir       string
	//rootDir           string
}

func (a *Album) String() string {
	return reflections.StructString(a, "newline", "\t\t")

}

func (a *Album) Add(filePath string) (*Disc, error) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "filePath", filePath)
	a.SetDestDir(filePath)

	track, err := NewTrack(filePath)
	if err != nil {
		slog.Error("NewTrack()|could not create new track", "filePath", filePath, "err", err)
		return nil, err

	}

	a.setName(track)

	slog.Log(context.TODO(), slogLevelArtist, "finished")
	return nil, nil

}

func (a *Album) addArtistName(artistName string) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "artistName", artistName)
	a.artistNames[artistName] = true

	slog.Log(context.TODO(), slogLevelArtist, "finished", "artistNames", a.artistNames)

}

func (a *Album) setDestDir(track *Track) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "track", track)

	nameToValue := track.nameToValue
	filePath := track.FilePath
	artistDir := filepath.Dir(filepath.Clean(filePath))
	albumBase := track.mp3.V2.Album()

	albumFolderExp := taggerConfig.AlbumFolderExp

	var originalAlbum string
	if value, ok := track.idToValue["TOAL"]; ok {
		originalAlbum = value

	}

	if track.MultiDiscAlbumName() != "" && originalAlbum != "" && !taggerConfig.TagAttr.NoDirectoryRename {
		albumFolderExp = strings.ReplaceAll(albumFolderExp, "{album}", "{original album}")

	}

	if taggerConfig.AlbumFolderExp != "" {
		albumBase = stringy.ReplaceUserString(albumFolderExp, nameToValue, stringy.TitleCase)

	}
	albumBase = filing.NormalizeFilename(albumBase)

	a.lastDestDir = a.DestDir
	a.DestDir = filepath.Join(artistDir, albumBase)
	a.baseDestDir = albumBase

	trackDir := filepath.Join(a.DestDir, track.Filename)
	track.SetDestDir(trackDir)

}

func (a *Album) SetDestDir(filePath string) {
	albumDir := filepath.Dir(filepath.Clean(filePath))
	if a.DestDir != albumDir {
		a.DestDir = albumDir

	}

}

func (a *Album) setName(track *Track) {
	albumName := track.mp3.V2.Album()
	if a.Name != albumName {
		a.Name = albumName

	}

}

func (a *Album) Discs() []*Disc {
	return a.discs

}

func (a *Album) IsDiscF(filePath string) (int, bool) {
	discNum, ok := a.filePathToDiscNum[filePath]

	return discNum - 1, ok

}

func (a *Album) IsDisc(discNum int) bool {
	if _, ok := a.numToDisc[discNum]; ok {
		return true

	}

	return false

}

func (a *Album) IsDisc2(discNum int) bool {
	if discNum <= 0 || discNum > len(a.discs) {
		return false

	}

	return true

}

func (a *Album) Disc(discNum int) *Disc {
	var disc *Disc
	if a.IsDisc(discNum) {
		disc = a.numToDisc[discNum]

	}

	return disc

}

func (a *Album) Disc2(discNum int) *Disc {
	var disc *Disc
	if a.IsDisc(discNum) {
		disc = a.discs[discNum-1]

	}

	return disc

}

func (a *Album) DiscF(filePath string) *Disc {
	var disc *Disc
	if discNum, ok := a.IsDiscF(filePath); ok {
		disc = a.discs[discNum]

	}

	return disc

}

func (a *Album) AddDisc(filePath string) (*Disc, error) {
	track, err := NewTrack(filePath)
	if err != nil {
		slog.Error("NewTrack()|could not create track", "filePath", filePath, "error", err)
		return nil, err

	}

	disc, err := a.addToDisc(track)

	return disc, err

}

func (a *Album) addToDisc2(track *Track) (*Disc, error) {
	discNum := track.DiscNumber()

	var disc *Disc
	if a.IsDisc(discNum) {
		disc = a.numToDisc[discNum]

	} else {
		var err error

		disc, err = NewDisc()
		if err != nil {
			return nil, err

		}

		a.numToDisc[discNum] = disc

	}

	a.filePathToDiscNum[track.FilePath] = discNum

	if disc.Number == 0 {
		disc.Number = discNum

	}
	disc.addToTracks(track)

	return disc, nil

}

func (a *Album) addToDisc(track *Track) (*Disc, error) {
	discNum := track.DiscNumber()

	var disc *Disc
	if a.IsDisc(discNum) {
		disc = a.Disc(discNum)

	} else {
		var err error
		disc, err = NewDisc()
		if err != nil {
			return nil, err

		}

		a.discs = append(a.discs, disc)

	}

	a.filePathToDiscNum[track.FilePath] = discNum
	if discNum > 1 {
		totalTracksLastDisc := a.discs[discNum-2].TotalTracks()
		trackNum, trackTot := track.mp3.V2.TrackNumber()
		trackNum -= totalTracksLastDisc
		var newTrackNum string = strconv.Itoa(trackNum)
		if trackTot > 0 {
			trackTot = disc.TotalTracks()
			newTrackNum += "/" + strconv.Itoa(trackTot)

		}

		track.SetTag("TRCK", newTrackNum)

	}

	if disc.Number == 0 {
		disc.Number = discNum

	}
	disc.addToTracks(track)

	return disc, nil

}

func (a *Album) DiscNum() int {
	return len(a.discs)

}

func (a *Album) TotalTracks() int {
	totalTracks := 0
	for _, disc := range a.numToDisc {
		totalTracks += disc.TotalTracks()

	}

	return totalTracks

}

func (a *Album) SetTotalTrackNumbers(setTotal bool) {
	for _, disc := range a.discs {
		disc.SetTrackNumbers(setTotal)

	}

}

// --------------------------------------------------------------------------------------------------
// func NewDisc(track *Track) (*Disc, error) {
func NewDisc() (*Disc, error) {
	disc := &Disc{}
	if err := reflections.InitializeStruct(disc); err != nil {
		return nil, err

	}

	return disc, nil

}

type Disc struct {
	DestDir            string   `json:"dest_dir"`
	Number             int      `json:"number"`
	Tracks             []*Track `json:"tracks"`
	trackNumToTrack    map[int]int
	filePathToTrackNum map[string]int
	lastDestDir        string
}

func (d *Disc) String() string {
	return reflections.StructString(d, "newline", "\t\t\t")

}

func (d *Disc) AddAll(dir string, tagAttrOpt ...*config.TagAttributes) error {
	if len(tagAttrOpt) > 0 && tagAttrOpt[0] != nil {
		taggerConfig.TagAttr = tagAttrOpt[0]

	}

	filename := ".*\\.mp3"
	filePaths := filing.LsEntryName(dir, filename, "add_dir")
	filenamesLen := len(filePaths)
	if filenamesLen == 0 {
		slog.Error("filing.LsEntryName()|nothing to process", "dir", dir, "filename", filename)
		return fmt.Errorf("nothing to process, dir: %s, filename: %s", dir, filename)

	}

	for _, filePath := range filePaths {
		d.Add(filePath)

	}

	return nil

}

func (d *Disc) Add(filePath string) (*Track, error) {
	track, err := NewTrack(filePath)
	if err != nil {
		slog.Error("NewTrack()|could not create new track", "filePath", filePath, "err", err)
		return nil, err

	}

	d.SetDestDir(track.FilePath)
	d.addToTracks(track)

	return track, nil

}

func (d *Disc) addToTracks(track *Track) {
	trackLen := len(d.Tracks)
	trackNum, _ := track.mp3.V2.TrackNumber()
	if trackNum != trackLen+1 {
		trackNum = trackLen
		track.SetTag("TRCK", strconv.Itoa(trackNum+1))

	}

	d.Tracks = append(d.Tracks, track)
	d.trackNumToTrack[trackNum] = trackLen
	d.filePathToTrackNum[track.FilePath] = trackNum

}

func (d *Disc) setDestDir(track *Track) {
	albumName := track.mp3.V2.Album()
	if d.lastDestDir != d.DestDir {
		d.lastDestDir = d.DestDir

	}

	if multiDiscAlbumName := track.MultiDiscAlbumName(); multiDiscAlbumName != "" {
		newChildFolder := filing.NormalizeFilename(multiDiscAlbumName, taggerConfig.FileReplaceSpaces)
		d.DestDir = filepath.Join(track.DestDir, newChildFolder)
		track.SetDestDir(filepath.Join(d.DestDir, track.Filename))

		if _, ok := track.idToValue["TOAL"]; !ok || (multiDiscAlbumName != albumName) {
			track.SetTag("TOAL", albumName)
			track.SetTag("TALB", multiDiscAlbumName)

		}

	} else {
		d.DestDir = track.DestDir

	}

}

func (d *Disc) SetDestDir(filePath string) {
	discDir := filepath.Dir(filepath.Clean(filePath))
	d.lastDestDir = d.DestDir
	d.DestDir = discDir

}

func (d *Disc) TrackF(filePath string) (*Track, error) {
	var track *Track
	if trackNum, ok := d.filePathToTrackNum[filePath]; ok {
		trackIndex := d.trackNumToTrack[trackNum]
		track = d.Tracks[trackIndex]

	} else {
		var err error
		track, err = NewTrack(filePath)
		if err != nil {
			return nil, err

		}

	}

	return track, nil

}

func (d *Disc) SetTrackNumbers(setTotal bool) {
	for _, track := range d.Tracks {
		trackNum, trackTot := track.mp3.V2.TrackNumber()
		newTrackNum := strconv.Itoa(trackNum)
		if setTotal {
			if trackTot != d.TotalTracks() {
				newTrackNum += "/" + strconv.Itoa(d.TotalTracks())

			}

		}

		track.SetTag("TRCK", newTrackNum)

	}

}
func (d *Disc) TotalTracks() int {
	return len(d.Tracks)

}

// --------------------------------------------------------------------------------------------------
func NewTrack(filePath string) (*Track, error) {
	destDir, filename := filepath.Split(filePath)

	mp3, err := MP3(filePath, true)
	if err != nil {
		slog.Error("MP3()|could not open file as mp3", "filePath", filePath, "err", err)
		return &Track{}, err

	}

	track := &Track{
		DestDir:          destDir,
		Filename:         filename,
		FilePath:         filePath,
		mp3:              mp3,
		OriginalFilePath: filePath,
		OriginalDir:      destDir,
		OriginalFilename: filename,
	}

	if err := reflections.InitializeStruct(track); err != nil {
		return nil, err

	}
	track.processFrames()

	return track, nil

}

func MP3(filePath string, initOpt ...bool) (*tunetag.MP3, error) {
	mp3, err := tunetag.OpenMP3(filePath)
	if err != nil {
		slog.Error("tunetag.OpenMP3()|could not open mp3", "filePath", filePath, "err", err)
		return nil, err

	}

	if len(initOpt) > 0 && initOpt[0] {
		slog.Debug("initialize mp3", "filePath", filePath, "initOpt", initOpt)
		initMp3(mp3)

	}

	return mp3, nil

}

func initMp3(mp3 *tunetag.MP3) {
	if mp3.V1 == nil {
		mp3.V1 = &id3v1.Tag{}

	}

	if mp3.V2 == nil {
		mp3.V2 = &id3v2.Tag{Version: id3v2.V23, Padding: id3v2.DefaultPadding}

	}

}

type Track struct {
	DestDir          string `json:"dest_dir"`
	Filename         string `json:"filename"`
	FilePath         string `json:"file_path"`
	OriginalDir      string `json:"original_dir"`
	OriginalFilename string `json:"original_filename"`
	OriginalFilePath string `json:"original_file_path"`
	PicProcessed     bool   `json:"pic_processed"`
	mp3              *tunetag.MP3
	idToValue        map[string]string
	nameToValue      map[string]string
	originalTags     map[string]string
	baseDestDir      string
	lastFilename     string
	lastFilePath     string
	lastDestDir      string
}

func (t *Track) String() string {
	return reflections.StructString(t, "newline", "\t\t\t\t")

}

func (t *Track) MP3() *tunetag.MP3 {
	return t.mp3

}

func (t *Track) SetDestDir(filePath string) {
	destDir, filename := filepath.Split(filePath)
	if t.OriginalFilePath == "" {
		t.OriginalDir = t.DestDir
		t.OriginalFilename = t.Filename
		t.OriginalFilePath = t.FilePath

	}

	t.lastDestDir = t.DestDir
	t.DestDir = destDir
	t.Filename = filename
	t.FilePath = filePath
	t.baseDestDir = filepath.Base(filepath.Clean(destDir))

}

func (t *Track) SetFilename(filename string) {
	t.lastFilename = t.Filename
	t.Filename = filename

}

func (t *Track) SetFilePath(filePath string) {
	t.lastFilePath = t.FilePath
	t.FilePath = filePath

}

func (t *Track) AddTag(id, value string) {
	if name, ok := config.IDToNameMod[id]; ok {
		if _, ok := t.originalTags[id]; !ok && t.idToValue[id] != "" {
			t.originalTags[id] = t.idToValue[id]

		}

		t.idToValue[id] = value
		t.nameToValue[name] = value

	} else { // If this happens and the id is valid, it needs to be added to the nameToIDMod struct
		slog.Error("invalid id, add to NameToIDMod struct", "id", id, "ok", ok)

	}

}

func (t *Track) Album() string {
	albumName := t.mp3.V2.Album()
	if originalAlbum, ok := t.idToValue["TOAL"]; ok && originalAlbum != "" {
		albumName = originalAlbum

	}

	return albumName

}

func (t *Track) MultiDiscAlbumName() string {
	var multiDiscAlbumName string
	discNum, discTot := t.mp3.V2.DiscNumber()
	if _, ok := t.idToValue["TOAL"]; ok && discTot > 1 {
		multiDiscAlbumName = t.mp3.V2.Album()

	} else if discTot > 1 {
		multiDiscAlbumAdd := taggerConfig.MultiDiscAlbumAdd
		if multiDiscAlbumAdd == "" {
			multiDiscAlbumAdd = "CD"

		}

		multiDiscAlbumName = fmt.Sprintf("%s (%s %d)", t.idToValue["TALB"], multiDiscAlbumAdd, discNum)

	}

	return multiDiscAlbumName

}

func (t *Track) SetTag(id, tagValue string, tagValuesOpt ...string) error {
	slog.Log(context.TODO(), slogLevelArtist, "started", "id", id, "tagValue", tagValue, "tagValuesOpt", tagValuesOpt)

	tagValue2 := ""
	noReplace := ""
	if len(tagValuesOpt) > 0 && tagValuesOpt[0] != "" {
		var found bool
		tagValuesOpt, found = filing.FindRegexOpt("no_replace", tagValuesOpt...)
		if found {
			noReplace = "no_replace"

		}

		if len(tagValuesOpt) > 0 {
			tagValue2 = tagValuesOpt[0]

		}

	}

	mp3 := t.mp3
	altTagType := config.NewAltTagType()
	if tagType, ok := altTagType.IDHasType(id); ok {
		t.AddTag(id, tagValue)

		if tagType == reflections.AnyType((*id3v2.PictureFrame)(nil)) && !t.PicProcessed { // APIC
			fmt.Printf("Artist.SetTag()|APIC\n")
			//if slices.Contains(tagValuesOpt, "no_replace") && len(mp3.V2.PictureFrames()) != 0 {
			if noReplace == "no_replace" && len(mp3.V2.PictureFrames()) != 0 {
				slog.Debug("SetTag()|no_replace", "tagValuesOpt", tagValuesOpt)
				return nil

			}

			if err := CreateCoverFile(t, tagValue); err != nil {
				return err

			}

			err := AddAPIC(mp3, t.DestDir, true)
			if err != nil {
				slog.Error("AddAPIC()|could not add picture frame", "err", err)
				return err

			}

			t.PicProcessed = true

		} else if tagType == reflections.AnyType((*id3v2.CommentFrame)(nil)) { // COMM
			addCommentFrame(mp3, tagValue)
			t.CopyToID1(id, tagValue)

		} else if tagType == reflections.AnyType((*id3v2.GenericFrame)(nil)) { // IPLS|GRP1|MVIN|MVNM|OWNE|PCST|POPM|SYLT
			addGenericFrame(mp3, id, tagValue)

		} else if tagType == reflections.AnyType((*id3v2.PrivFrame)(nil)) { // PRIV
			addPrivFrame(mp3, id, tagValue, tagValue2)

		} else if tagType == reflections.AnyType((*id3v2.UFIDFrame)(nil)) { // UFID
			addUFIDFrame(mp3, id, tagValue, tagValue2)

		} else if tagType == reflections.AnyType((*id3v2.UnsynchronisedLyricsFrame)(nil)) { // USLT
			addUnsynchronisedLyricsFrame(mp3, id, tagValue, tagValue2)

		} else if tagType == reflections.AnyType((*id3v2.URLFrame)(nil)) { // WCOP|WFED|WOAF|WOAR|WORS|WPAY|WPUB
			addURLFrame(mp3, id, tagValue)

		} else if tagType == reflections.AnyType((*id3v2.UserTextFrame)(nil)) { // TXXX
			addUserTextFrame(mp3, tagValue, tagValue2)

		} else if tagType == reflections.AnyType((*id3v2.UserURLFrame)(nil)) { // WXXX
			addUserURLFrame(mp3, id, tagValue, tagValue2)

		}

	} else {
		slog.Debug("SetText()|setting tag", "id", id, "tagValue", tagValue)
		mp3.V2.SetText(id, tagValue)
		t.AddTag(id, tagValue)

		if config.RequiredTags.IsCopyTo(id) { // Copy this value to other fields
			for _, field := range config.RequiredTags.CopyToFields[id] {
				slog.Debug("SetText()|setting tag 2", "field", field, "tagValue", tagValue)
				mp3.V2.SetText(field, tagValue)
				t.AddTag(field, tagValue)

			}
		}

		t.CopyToID1(id, tagValue)

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished")
	return nil

}

// copyToID1 copies an mp3.V2 tag to mp3.V1 if applicable (i.e. there is a corresponding tag specified in requiredTags)
func (t *Track) CopyToID1(id, value string) {
	slog.Log(context.TODO(), slogLevelArtist, "started", "id", id, "value", value)

	regexCurly := regexp.MustCompile("’")
	if _, ok := config.RequiredTags.Is2In1(id); ok {
		newValue := value
		if id == "COMM" && value != "" { // Comment
			newValue = t.mp3.V2.Comment()

		} else if id == "TCON" {
			if value == "" {
				newValue = strconv.Itoa(255)

			} else {
				newValue = findV1Genre(value)

			}

		}
		newValue = regexCurly.ReplaceAllString(newValue, "'")

		v1Field := config.RequiredTags.IdV2ToIDV1[id]
		slog.Debug("copying mp3.V2 tag to mp3.V1", "id", id, "value", value, "v1Field", v1Field, "newValue", newValue)
		reflections.SetStructField(t.mp3.V1, v1Field, newValue)

	}

	slog.Log(context.TODO(), slogLevelArtist, "finished")

}

func (t *Track) DiscNumber() int {
	discNum, _ := t.mp3.V2.DiscNumber()
	if discNum == 0 {
		discNum = 1

	}

	return discNum

}

func (t *Track) Value(id string) string {
	if value, ok := t.idToValue[id]; ok {
		return value

	} else {
		slog.Error("t.idToValue()|invalid", "id", id, "ok", ok)

	}

	return ""

}

func (t *Track) Write() error {
	if err := t.mp3.V1.WriteFile(t.OriginalFilePath); err != nil {
		return err

	}

	if err := t.mp3.V2.WriteFile(t.OriginalFilePath); err != nil {
		return err

	}

	return nil

}

func (t *Track) processFrames() {
	// Don't want both Date and Year. Remove TYER now so it doesn't mess up the Frames
	year := t.mp3.V2.Year()
	if year != 0 {
		t.mp3.V2.RemoveFrames("TYER")

	}

	for _, f := range t.mp3.V2.Frames {
		fID := f.ID()
		fValue := fmt.Sprintf("%s", f)
		fName := fID
		if name, ok := config.IDToNameMod[fID]; ok {
			fName = name

		}

		if tagType, ok := altTagType.IDHasType(fID); ok {
			if tagType == reflections.AnyType((*id3v2.PictureFrame)(nil)) { // APIC
				fValue = "APIC"

			} else if tagType == reflections.AnyType((*id3v2.CommentFrame)(nil)) { // COMM
				fValue = f.(*id3v2.CommentFrame).Text

			} else if tagType == reflections.AnyType((*id3v2.GenericFrame)(nil)) { // IPLS|GRP1|MVIN|MVNM|OWNE|PCST|POPM|SYLT
				fValue = string(f.(*id3v2.GenericFrame).Body)

			} else if tagType == reflections.AnyType((*id3v2.UserTextFrame)(nil)) { // TXXX
				fValue = f.(*id3v2.UserTextFrame).Value

			} else if tagType == reflections.AnyType((*id3v2.URLFrame)(nil)) { // WCOP|WFED|WOAF|WOAR|WORS|WPAY|WPUB
				fValue = f.(*id3v2.URLFrame).URL

			} else if tagType == reflections.AnyType((*id3v2.PrivFrame)(nil)) { // PRIV
				fValue = string(f.(*id3v2.PrivFrame).Data)

			} else if tagType == reflections.AnyType((*id3v2.UFIDFrame)(nil)) { // UFID
				fValue = string(f.(*id3v2.UFIDFrame).Identifier)

			} else if tagType == reflections.AnyType((*id3v2.UnsynchronisedLyricsFrame)(nil)) { // USLT
				fValue = f.(*id3v2.UnsynchronisedLyricsFrame).Text

			} else if tagType == reflections.AnyType((*id3v2.UserURLFrame)(nil)) { // WXXX
				fValue = f.(*id3v2.UserURLFrame).URL

			}

		} else {
			if fID == "TRCK" {
				t.idToValue[fID+"TOT"] = fValue
				t.nameToValue[fName+" Total"] = fValue
				trackNum, _ := t.mp3.V2.TrackNumber()
				fValue = strconv.Itoa(trackNum)
				t.SetTag(fID, fValue)

			} else if fID == "TPOS" && taggerConfig.RemoveDisc1Of1 {
				discNum, discTot := t.mp3.V2.DiscNumber()
				if discNum == 1 && (discTot == 1 || discTot == 0) {
					t.mp3.V2.RemoveFrames("TPOS")

				}

			} else {
				curlyRe := regexp.MustCompile("’")
				if curlyRe.MatchString(fValue) {
					fValue = curlyRe.ReplaceAllString(fValue, "'")
					t.SetTag(fID, fValue)
					slog.Debug("ReplaceAllString()|removed curly quote", "fID", fID, "fValue", fValue)

				}

			}

			if taggerConfig.RemoveBonusTrack && fID == "TIT2" {
				bonusTrackRe := " \\([Bb]onus [Tt]rack\\)"
				regex := regexp.MustCompile(bonusTrackRe)
				if regex.MatchString(fValue) {
					fValue = regex.ReplaceAllString(fValue, "")

				}

			}

		}

		t.idToValue[fID] = fValue
		t.nameToValue[fName] = fValue

		if config.RequiredTags.IsCopyTo(fID) { // Copy this value to other fields
			for _, field := range config.RequiredTags.CopyToFields[fID] {
				slog.Debug("SetText()", "field", field, "fValue", fValue)
				t.mp3.V2.SetText(field, fValue)
				t.AddTag(field, fValue)

			}

		}

		t.CopyToID1(fID, fValue)

	}

	if year != 0 {
		t.idToValue["TYER"], t.nameToValue["Year"] = strconv.Itoa(year), strconv.Itoa(year)
		t.SetTag("TDRC", strconv.Itoa(year))

	}

}

func (t *Track) processPreConditions() error {
	slog.Log(context.TODO(), slogLevelNoOp, "started")
	idToPreConditions, err := taggerConfig.TagAttr.SplitPreConditions()
	if err != nil {
		return err

	}

	for id, preConditions := range idToPreConditions {
		for _, preCondition := range preConditions {
			rangeRegexp := regexp.MustCompile("^[0-9]+-[0-9]+$")
			if id == "TRCK" && rangeRegexp.MatchString(preCondition.RegEx) {
				preConditionParts := strings.Split(preCondition.RegEx, "-")
				range1, _ := strconv.Atoi(preConditionParts[0])
				range2, _ := strconv.Atoi(preConditionParts[1])
				trackNum, _ := t.mp3.V2.TrackNumber()
				if trackNum >= range1 && trackNum <= range2 {
					t.SetTag(preCondition.DestTag, preCondition.DestTagValue)

				} else if preCondition.DestTagElseValue != nil && *preCondition.DestTagElseValue != "" {
					t.SetTag(preCondition.DestTag, *preCondition.DestTagElseValue)

				}

			} else {
				searchRegexp, err := regexp.Compile(preCondition.RegEx)
				if err != nil {
					slog.Error("regexp.Compile()|could not compile regex", "preCondition", preCondition)
					return err

				}

				checkValue := t.Value(id)
				if id == "File" {
					checkValue = t.Filename

				}
				slog.Debug("POES", "f.Filename", t.Filename)

				if searchRegexp.MatchString(checkValue) {
					t.SetTag(preCondition.DestTag, preCondition.DestTagValue)

				} else if preCondition.DestTagElseValue != nil && *preCondition.DestTagElseValue != "" {
					t.SetTag(preCondition.DestTag, *preCondition.DestTagElseValue)

				}

			}

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")
	return nil

}

func (t *Track) processReplacements(isPre bool) error {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "isPre", isPre)
	var idToReplacements map[string][]config.Replacements
	var err error
	if isPre {
		idToReplacements, err = taggerConfig.TagAttr.SplitPreReplacements()

	} else {
		idToReplacements, err = taggerConfig.TagAttr.SplitReplacements()

	}
	if err != nil {
		slog.Error("SplitReplacements()|could not process Replacements", "err", err)
		return err

	}

	for id, replacements := range idToReplacements {
		if id == "APIC" {
			slog.Error("input|cannot perform a replacement on the APIC (cover art) frame")
			continue

		} else if !isPre && id == "TALB" {
			slog.Error("input|cannot perform a replacement on the TALB (Album) frame")
			continue

		}

		if value, ok := t.idToValue[id]; ok {
			for _, replacement := range replacements {
				replaceRe, replaceExp := replacement.RegEx, replacement.Replace
				replaceRegexp, err := regexp.Compile(replaceRe)
				if err != nil {
					slog.Error("regexp.Compile()|problem with regular expression", "id", id, "replaceRe", replaceRe, "replaceExp", replaceExp)
					return err

				}
				value = replaceRegexp.ReplaceAllString(value, replaceExp)

				slog.Debug("ReplaceAllString()|replaced", "id", id, "value", value, "replaceRe", replaceRe, "replaceExp", replaceExp)
				t.SetTag(id, value)
				t.idToValue[id] = value

			}

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")
	return nil

}

func (t *Track) Rename() error {
	slog.Log(context.TODO(), slogLevelNoOp, "started")

	if taggerConfig.FileRenameExp != "" {
		ext := filepath.Ext(t.Filename)
		newFilenameExp := taggerConfig.FileRenameExp + ext
		newFilename := stringy.ReplaceUserString(newFilenameExp, t.nameToValue, stringy.TitleCase)
		normalizedFilename := filing.NormalizeFilename(newFilename, taggerConfig.FileReplaceSpaces)

		if newFilename != t.OriginalFilename {
			newFilePath := filepath.Join(t.DestDir, normalizedFilename)
			slog.Info("renaming file", "OriginalFilename", t.OriginalFilename, "normalizedFilename", normalizedFilename, "OriginalFilePath", t.OriginalFilePath, "DestDir", t.DestDir)
			os.MkdirAll(t.DestDir, filing.DirPerm)
			if err := os.Rename(t.OriginalFilePath, newFilePath); err != nil {
				slog.Error("os.Rename()|could not rename file", "OriginalFilePath", t.OriginalFilePath, "newFilePath", newFilePath)
				return err

			}

			t.SetFilename(normalizedFilename)
			t.SetFilePath(newFilePath)

		}

	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished")
	return nil

}
