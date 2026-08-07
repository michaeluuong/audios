package mp3

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/michaeluuong/utilize/filing"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()

	os.Exit(code)

}

var testDir string

var eaglesCoverArt, eaglesDir, eaglesAlbumDir string
var eaglesFiles []string

var vanaDir, vanaAlbumDir string
var vanaFiles []string

func setup() {
	var err error
	testDir, err = filepath.Abs("./TEST_FILES")
	if err != nil {
		panic(err)

	}

	eaglesDir = filepath.Join(testDir, "Beagles-One_Of_These_Frights_(Deluxe_Edition)-WEB-2026")
	eaglesAlbumDir = filepath.Join(eaglesDir, "Beagles-One_Of_These_Frights_(Deluxe_Edition)-WEB-2026") + "/"
	eaglesCoverArt = filepath.Join(eaglesAlbumDir, "00-beagles-one_of_these_frights_(deluxe_edition)-web-2026.png")
	//eaglesCoverArt = "00-beagles-one_of_these_frights_(deluxe_edition)-web-2026.png"
	eaglesFiles = filing.LsEntryName(eaglesAlbumDir, ".*\\.mp3", "add_dir")

	vanaDir = filepath.Join(testDir, "Vana-In_Your_Name_bw_Pray-Web-2026")
	vanaAlbumDir = filepath.Join(vanaDir, "Vana-In_Your_Name_bw_Pray-Web-2026") + "/"
	vanaFiles = filing.LsEntryName(vanaAlbumDir, ".*\\.mp3", "add_dir")

}

func teardown() {

}

// go test -benchtime=1s -bench . -cpuprofile cpu.prof
// go tool pprof cpu.prof
func BenchmarkReflections(b *testing.B) {

}

func TestNewArtist_main(t *testing.T) {
	var artist *Artist
	var err error

	//--------------------------------------------------------------------------------------------
	// Happy
	artist, err = NewArtist()
	assert.Nil(t, err)
	assert.NotNil(t, artist.nameToAlbum)
	assert.NotEmpty(t, artist.nameToAlbum)
	assert.NotNil(t, artist.albumNamesToAlbumNum)
	assert.Empty(t, artist.albumNamesToAlbumNum)
	assert.NotNil(t, artist.originalFilePathToAlbumName)
	assert.Empty(t, artist.originalFilePathToAlbumName)

	artist.albumNamesToAlbumNum["Album1"] = 1
	artist, err = NewArtist(artist)
	assert.Equal(t, artist.albumNamesToAlbumNum["Album1"], 1)

	//--------------------------------------------------------------------------------------------
	// Error
	artist, err = NewArtist(nil)
	assert.ErrorContains(t, err, "must be a pointer to a struct, kind is ")

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestNewArtist_rest(t *testing.T) {

	//--------------------------------------------------------------------------------------------
	artist, err := NewArtist()
	assert.Nil(t, err)
	for _, filename := range vanaFiles {
		artist.Add(filename)

	}
	fmt.Printf("artist: %s\n", artist)

	// album
	albums := artist.albums
	assert.Equal(t, "In Your Name b/w Pray Web", albums[0].Name)
	assert.Equal(t, vanaAlbumDir, albums[0].DestDir)
	assert.Equal(t, albums[0].artistNames, map[string]bool{"Vana": true})
	assert.Equal(t, 1, albums[0].filePathToDiscNum[vanaFiles[0]])
	assert.Equal(t, 1, albums[0].filePathToDiscNum[vanaFiles[1]])

	// disc
	discs := albums[0].discs
	assert.Equal(t, 1, len(discs))
	assert.Equal(t, 1, discs[0].filePathToTrackNum[vanaFiles[0]])
	assert.Equal(t, 2, discs[0].filePathToTrackNum[vanaFiles[1]])

	// track
	track := discs[0].Tracks
	assert.Equal(t, 1, len(discs))
	assert.Equal(t, 2, len(discs[0].Tracks))
	fmt.Printf("tracks: %v\n", discs[0].Tracks)

	// Track 1
	mp3, err := MP3(vanaFiles[0])
	frames := mp3.V2.Frames
	assert.Equal(t, mp3, discs[0].Tracks[0].mp3)
	assert.Equal(t, vanaAlbumDir, discs[0].Tracks[1].DestDir)
	assert.Equal(t, "01-vana-in_your_name.mp3", discs[0].Tracks[0].Filename)
	assert.Equal(t, vanaFiles[0], discs[0].Tracks[0].FilePath)

	// Check individual frames
	for _, frame := range frames {
		fID := frame.ID()
		fValue := fmt.Sprintf("%s", frame)
		assert.Equal(t, fValue, track[0].Value(fID))
	}

	// Track 2
	mp3_2, err := MP3(vanaFiles[1])
	frames_2 := mp3_2.V2.Frames
	assert.Equal(t, mp3_2, discs[0].Tracks[1].mp3)
	assert.Equal(t, vanaAlbumDir, discs[0].Tracks[1].DestDir)
	assert.Equal(t, "02-vana-pray.mp3", discs[0].Tracks[1].Filename)
	assert.Equal(t, vanaFiles[1], discs[0].Tracks[1].FilePath)

	// Check individual frames
	for _, frame := range frames_2 {
		fID := frame.ID()
		fValue := fmt.Sprintf("%s", frame)
		assert.Equal(t, fValue, track[1].Value(fID))
	}

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestDirs_main(t *testing.T) {
	artist, err := NewArtist()
	assert.Nil(t, err)

	for _, file := range eaglesFiles {
		artist.Add(file)

	}
	fmt.Printf("artist: %s\b", artist)

	//--------------------------------------------------------------------------------------------

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestString_main(t *testing.T) {

	//--------------------------------------------------------------------------------------------
	// Happy
	// Just checking that Stringer has been implemented
	artistString := fmt.Sprintf("artist: %s", &Artist{})
	assert.NotNil(t, artistString)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestAdd_main(t *testing.T) {
	var artist *Artist
	var err error

	//--------------------------------------------------------------------------------------------
	// Happy
	artist, err = NewArtist()
	assert.Nil(t, err)

	artist.Add(eaglesFiles[0])
	fmt.Printf("artist: %s\n", artist)

	assert.Equal(t, eaglesDir, artist.DestDir)
	assert.Equal(t, "Beagles", artist.Name)

	album := artist.albums[0]
	assert.Equal(t, "One Of These Frights (Deluxe Edition)", album.Name)
	assert.Empty(t, album.Cover)
	assert.Equal(t, eaglesAlbumDir, album.DestDir)
	assert.True(t, album.artistNames["Beagles"])

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestTrack_main(t *testing.T) {
	var track *Track
	var err error

	//--------------------------------------------------------------------------------------------
	// Happy
	track, err = NewTrack(eaglesFiles[0])
	mp3 := track.MP3()
	assert.Nil(t, err)
	assert.Equal(t, eaglesAlbumDir, track.DestDir)
	assert.Equal(t, filepath.Base(eaglesFiles[0]), track.Filename)
	assert.Equal(t, eaglesFiles[0], track.FilePath)
	assert.Equal(t, mp3.V2.Artist(), track.Value("TPE1"))
	assert.Equal(t, mp3.V2.Album(), track.Value("TALB"))
	assert.Equal(t, mp3.V2.Title(), track.Value("TIT2"))

	trackNum, _ := mp3.V2.TrackNumber()
	assert.Equal(t, fmt.Sprintf("%02d", trackNum), track.Value("TRCK"))
	assert.Equal(t, fmt.Sprintf("%d", mp3.V2.Year()), track.Value("TDRC"))
	fmt.Printf("track: %v\n", track)

	//--------------------------------------------------------------------------------------------
	// Error
	_, err = NewTrack("")
	fmt.Printf("err: %v\n", err)
	assert.Error(t, err, "open : no such file or directory")

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestSetTag_main(t *testing.T) {
	track, err := NewTrack(eaglesFiles[0])
	mp3 := track.MP3()

	//--------------------------------------------------------------------------------------------
	// Set tag types
	// Picture
	assert.Empty(t, track.mp3.V2.PictureFrames())
	fmt.Printf("\n\neaglesCoverArt: %v\n\n", eaglesCoverArt)
	err = track.SetTag("APIC", eaglesCoverArt)
	assert.Nil(t, err)
	assert.NotEmpty(t, track.mp3.V2.PictureFrames())
	track.mp3.V2.RemoveFrames("APIC")
	os.Remove(filepath.Join(eaglesAlbumDir, "cover68.png"))

	assert.Empty(t, track.mp3.V2.PictureFrames())
	err = track.SetTag("APIC", "https://www.bil-jac.com/wp-content/uploads/2024/12/beagle2-184102750.webp")
	assert.Nil(t, err)
	assert.NotEmpty(t, track.mp3.V2.PictureFrames())
	os.Remove(filepath.Join(eaglesAlbumDir, "cover69.jpg"))

	// Album Artist
	err = track.SetTag("TPE2", "Beagles Album Artist")
	assert.Nil(t, err)
	assert.Equal(t, mp3.V2.AlbumArtist(), track.Value("TPE2"))

	// Comment/CommentFrame
	err = track.SetTag("COMM", "New Comment Test")
	assert.Nil(t, err)
	assert.Equal(t, mp3.V2.Comment(), track.Value("COMM"))
	assert.Equal(t, mp3.V2.Comment(), mp3.V1.Comment)

	// Arranger/GenericFrame
	err = track.SetTag("IPLS", "New Arranger")
	assert.Nil(t, err)
	assert.Equal(t, "New Arranger", track.Value("IPLS"))

	// User/UserTextFrame
	err = track.SetTag("TXXX", "New User", "UserDescription")
	assert.Nil(t, err)
	assert.Equal(t, "New User", track.Value("TXXX"))

	// Copyright Information/URLFrame
	err = track.SetTag("WCOP", "http://WCOP")
	assert.Nil(t, err)
	assert.Equal(t, "http://WCOP", track.Value("WCOP"))

	// Private Frame/PrivFrame
	err = track.SetTag("PRIV", "Private Frame", "IOU")
	assert.Nil(t, err)
	assert.Equal(t, "Private Frame", track.Value("PRIV"))

	// Unique File Identifier/UFIDFrame
	err = track.SetTag("UFID", "Unique File Identifier", "IOU")
	assert.Nil(t, err)
	assert.Equal(t, "Unique File Identifier", track.Value("UFID"))

	// Unsynchronised Lyrics Frame/UnsynchronisedLyricsFrame
	err = track.SetTag("USLT", "Unsynchronised Lyrics Frame", "Non-descript")
	assert.Nil(t, err)
	assert.Equal(t, "Unsynchronised Lyrics Frame", track.Value("USLT"))

	// User Defined URL Link Frame/UserURLFrame
	err = track.SetTag("WXXX", "http://WXXX", "Non-descript")
	assert.Nil(t, err)
	assert.Equal(t, "http://WXXX", track.Value("WXXX"))

	err = track.Write()

	//--------------------------------------------------------------------------------------------
	// Cleanup
	cleanUpTrack, err := NewTrack(eaglesFiles[0])
	cleanUpTrack.mp3.V2.RemoveFrames("APIC")
	cleanUpTrack.mp3.V2.RemoveFrames("TPE2")
	cleanUpTrack.mp3.V2.RemoveFrames("COMM")
	cleanUpTrack.mp3.V2.RemoveFrames("IPLS")
	cleanUpTrack.mp3.V2.RemoveFrames("TXXX")
	cleanUpTrack.mp3.V2.RemoveFrames("WCOP")
	cleanUpTrack.mp3.V2.RemoveFrames("PRIV")
	cleanUpTrack.mp3.V2.RemoveFrames("UFID")
	cleanUpTrack.mp3.V2.RemoveFrames("USLT")
	cleanUpTrack.mp3.V2.RemoveFrames("WXXX")
	err = cleanUpTrack.Write()

}

//func Test_main(t *testing.T) {}
