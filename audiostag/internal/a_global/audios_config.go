package a_global

import (
	"os"

	archiver "github.com/michaeluuong/audios/audiostag/internal/app/core/archiver/config"
	mp3Config "github.com/michaeluuong/audios/audiostag/internal/app/core/mp3/config"
	"github.com/michaeluuong/audios/audiostag/internal/pkg/config"
	"github.com/michaeluuong/utilize/reflections"
)

const (
	audiostagConfigFilename = "audiostag_cfg.json"
)

type AudiostagConfig struct {
	CLI     *CLIConfig              `json:"cli"`
	Archive *archiver.ArchiveConfig `json:"archive"`
	MP3     *mp3Config.MP3Config    `json:"mp3"`
	config.DefaultConfigger
}

func (a *AudiostagConfig) String() string {
	return reflections.StructString(a, "newline", "tab")

}

func (a *AudiostagConfig) SetupConfig() error {
	if err := a.DefaultConfigger.DefaultSetupConfig(a); err != nil {
		return err

	}

	return nil

}

func NewAudiostagConfig(dirNameOpt ...string) *AudiostagConfig {
	dirName := os.Getenv("PROGNAME")
	if len(dirNameOpt) >= 0 && dirNameOpt[0] != "" {
		dirName = dirNameOpt[0]

	}

	newAudiostagConfig := &AudiostagConfig{
		DefaultConfigger: config.DefaultConfigger{
			DefaultConfigFilename: audiostagConfigFilename,
			DirName:               dirName,
			DefaultParentValues: &AudiostagConfig{
				/*MP3: &mp3.MP3Config{
					CoverFilenameNoExt: "cover",
					CoverSleepTime:     0,
					FeaturingFix:       "feat",
					FeaturingParen:     true,
					RemoveGenre:        true,
					FileRenameExp:      "%02s{track number}. %{title}",
					AlbumFolderExp:     "(%{year}) %{album}",
					ArtistFolderExp:    "%{artist}",
					PlaylistExp:        "%{artist} - %{album}",
					MaxFileLimit:       0,
					MultiDiscAlbumAdd:  "CD",
					ShowTags:           true,
					SplitDisc:          true,
				},*/
				CLI:     &CLIConfig{},
				MP3:     mp3Config.NewMP3Config(),
				Archive: archiver.NewArchiveConfig(),
				/*Archive: &archiver.ArchiveConfig{
					CreateTopLevelDir: true,
					ExcludeFileRegex:  "\\.(nfo|sfv|m3u)$|.*proof.*\\.(jpg|png)",
					ExtractConcurrent: 1,
					ExtractRename: map[string]string{
						".jpg":    "cover",
						".png":    "cover",
						".gif":    "cover",
						"exclude": ".*proof.*",
					},
					TrashArchive: false,
				},*/
			},
			IsPanic: true,
		},
	}

	return newAudiostagConfig

}

func AudiostagConfigMan(dirNameOpt ...string) *AudiostagConfig {
	dirName := ""
	if len(dirNameOpt) > 0 {
		dirName = dirNameOpt[0]

	}

	configMan := config.GetConfigManInstance()
	anyAudiostagConfig := configMan.ConfigPtr(new(AudiostagConfig))
	var audiostagConfig *AudiostagConfig
	if *anyAudiostagConfig == nil {
		audiostagConfig = NewAudiostagConfig(dirName)
		configMan.AddConfig(audiostagConfig)

	} else {
		audiostagConfig, _ = config.AssertConfigger[AudiostagConfig](anyAudiostagConfig)

	}

	return audiostagConfig

}

func FromConfigMan() (*AudiostagConfig, bool) {
	configMan := config.GetConfigManInstance()
	anyAudiostagConfig := configMan.ConfigPtr(new(AudiostagConfig))

	return config.AssertConfigger[AudiostagConfig](anyAudiostagConfig)

}
