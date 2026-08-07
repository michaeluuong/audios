package mp3

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kgiannakakis/mp3duration/src/mp3duration"
	"github.com/lizc2003/audioduration"

	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/reflections"
	"github.com/michaeluuong/utilize/stringy"
)

type PlaylistSong struct {
	Artist   string
	Album    string
	Duration float64
	Filename string
	Title    string
}

type Playlist struct {
	VariousArtists bool
	PlaylistSongs  []PlaylistSong
}

func NewPlaylist() *Playlist {
	return &Playlist{}

}

func (p *Playlist) AddPlaylistSong(artist, title, filePath string) {
	_, filename := filepath.Split(filePath)
	duration, err := p.Duration(filePath)
	if err != nil {
		slog.Error("Duration()", "err", err)
		duration = -1

	}

	p.PlaylistSongs = append(p.PlaylistSongs, PlaylistSong{
		Artist:   artist,
		Title:    title,
		Filename: filename,
		Duration: duration,
	})

}

func (p *Playlist) AddPlaylistSongStruct(pls *PlaylistSong) {
	p.PlaylistSongs = append(p.PlaylistSongs, *pls)

}

func (p *Playlist) ClearPlaylist() {
	p.PlaylistSongs = p.PlaylistSongs[:0]

}

func DurationSlow(filePath string) (float64, error) {
	duration, err := mp3duration.Calculate(filePath)
	if err != nil {
		return 0, err
	}

	return duration, nil

}

func (p *Playlist) Duration(filePath string) (float64, error) {
	file, _ := os.Open(filePath)
	defer file.Close()

	duration, _ := audioduration.Mp3(file) // returns float64 in seconds

	return duration, nil

}

func (p *Playlist) ScratchPlaylist(disc *Disc, playlistNameOpt ...string) error {
	if len(disc.Tracks) <= 20 {
		slog.Debug("filing.LsEntryName*()|files to add to playlist", "Tracks", disc.Tracks)

	}

	var nameToValue map[string]string
	for _, track := range disc.Tracks {
		if nameToValue == nil {
			nameToValue = track.nameToValue

		}

		filePath := filepath.Join(disc.DestDir, track.Filename)
		duration, err := p.Duration(filePath)
		playlistSong, err := SongInfoToPlaylist(filePath, duration)
		if err != nil {
			slog.Error("SongInfoToPlaylist()|could not get playlistSong from file", "filePath", filePath, "duration", duration, "err", err)
			return err

		}

		slog.Info("adding song to playlist", "playlistSong", playlistSong)
		p.AddPlaylistSongStruct(playlistSong)

	}

	var playlistFilename string = taggerConfig.PlaylistExp
	if len(playlistNameOpt) > 0 && playlistNameOpt[0] != "" {
		playlistFilename = playlistNameOpt[0]

	} else {
		if p.VariousArtists {
			playlistFilename = "%{album}"

		} else if playlistFilename == "" {
			playlistFilename = "%{artist} - %{album}"

		}

	}

	return p.CreatePlaylist(disc.DestDir, nameToValue, playlistFilename)

}

func SongInfoToPlaylist(filePath string, duration float64) (*PlaylistSong, error) {
	slog.Log(context.TODO(), slogLevelNoOp, "started", "filePath", filePath, "duration", duration)

	mp3, err := MP3(filePath)
	if err != nil {
		return &PlaylistSong{}, err

	}

	_, filename := filepath.Split(filePath)
	playlistSong := &PlaylistSong{
		Artist:   mp3.V2.Artist(),
		Album:    mp3.V2.Album(),
		Duration: duration,
		Title:    mp3.V2.Title(),
		Filename: filename,
	}

	slog.Log(context.TODO(), slogLevelNoOp, "finished", "playlistSong", playlistSong)
	return playlistSong, nil

}

func playlistSongToMap(playlistSong PlaylistSong) map[string]string {
	fieldToValue := make(map[string]string)

	v := reflect.ValueOf(playlistSong)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !reflections.IsStruct(playlistSong) {
		return fieldToValue

	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue

		}

		fieldToValue[field.Name] = fmt.Sprintf("%s", v.Field(i).Interface())

	}

	return fieldToValue

}

func (p *Playlist) CreatePlaylist(dir string, namesToValues map[string]string, playlistFilename string) error {
	playlistSongLen := len(p.PlaylistSongs)
	if playlistSongLen == 0 {
		return errors.New("there are no songs for the playlist")

	}

	playlistFilename = stringy.ReplaceUserString(playlistFilename, namesToValues, stringy.TitleCase) + ".m3u"
	playlistFilename = filing.NormalizeFilename(playlistFilename, taggerConfig.FileReplaceSpaces)

	filename := filepath.Join(dir, playlistFilename)
	file, err := os.Create(filename)
	if err != nil {
		slog.Error("os.Create()|could not create playlist file", "filename", filename, "err", err)
		return err

	}
	defer file.Close()

	slog.Info("creating playlist with values from the first file in directory", "playlistSong[0]", p.PlaylistSongs[0], "filename", filename)

	writer := bufio.NewWriter(file)
	writer.WriteString("#EXTM3U\n")
	for _, playlistSong := range p.PlaylistSongs {
		extinf := fmt.Sprintf("#EXTINF:%d,%s - %s\n", int(math.Floor(playlistSong.Duration)), playlistSong.Artist, playlistSong.Title)
		writer.WriteString(extinf)

		songFilename := fmt.Sprintf("%s\n", playlistSong.Filename)
		writer.WriteString(songFilename)

	}

	writer.Flush()

	return nil

}
