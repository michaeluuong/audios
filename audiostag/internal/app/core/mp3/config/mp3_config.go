package config

import "github.com/michaeluuong/utilize/reflections"

const (
	archiveConfigFilename = "mp3_cfg.json"
	artworkRegexDefault   = "^(cover|folder).*\\.(jpg|png|gif)$"
)

//var taggerConfig = MP3ConfigMan()

type MP3Config struct {
	AlbumFolderExp     string         `json:"album_folder_exp"`
	ArtistFolderExp    string         `json:"artist_folder_exp"`
	ArtistToTitle      string         `json:"artist_to_title"`
	CoverFilenameNoExt string         `json:"cover_filename_no_ext"`
	CoverSize          string         `json:"cover_size"`
	CoverSleepTime     int            `json:"cover_sleep_time"`
	FeaturingFix       string         `json:"featuring_fix"`
	FeaturingParen     bool           `json:"featuring_paren"`
	FileReplaceSpaces  bool           `json:"file_replace_spaces"`
	FileRenameExp      string         `json:"file_rename_exp"`
	MaxFileLimit       int            `json:"max_file_limit"`
	MultiDiscAlbumAdd  string         `json:"multi_disc_album_add"`
	PlaylistExp        string         `json:"playlist_exp"`
	RemoveDisc1Of1     bool           `json:"remove_disc_1_of_1"`
	RemoveBonusTrack   bool           `json:"remove_bonus_track"`
	RemoveGenre        bool           `json:"remove_genre"`
	ShowTags           bool           `json:"show_tags"`
	SplitDisc          bool           `json:"split_disc"`
	TagAttr            *TagAttributes `json:"tag_attributes"`
	TagReplacements    string         `json:"tag_replacements"`
	TotalTracks        bool           `json:"total_tracks"`
}

func (m *MP3Config) String() string {
	return reflections.StructString(m)

}

func (m *MP3Config) ArtworkRegex() string {
	artworkRegex := artworkRegexDefault
	if m.CoverFilenameNoExt != "" {
		artworkRegex = "^(" + m.CoverFilenameNoExt + ").*\\.(jpg|png|gif|)$"

	}

	return artworkRegex

}

func NewMP3Config() *MP3Config {
	return &MP3Config{
		CoverFilenameNoExt: "cover",
		CoverSize:          "T1200",
		CoverSleepTime:     0,
		FeaturingFix:       "feat",
		FeaturingParen:     true,
		RemoveBonusTrack:   true,
		RemoveGenre:        true,
		FileRenameExp:      "%02s{track number}. %{title}",
		AlbumFolderExp:     "(%{year}) %{album}",
		ArtistFolderExp:    "%{artist}",
		PlaylistExp:        "%{artist} - %{album}",
		MaxFileLimit:       0,
		MultiDiscAlbumAdd:  "CD",
		ShowTags:           true,
		SplitDisc:          true,
	}

}
