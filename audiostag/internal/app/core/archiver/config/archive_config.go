package archiver

import "github.com/michaeluuong/utilize/reflections"

const (
	archiveConfigFilename = "archive_cfg.json"
)

type ArchiveConfig struct {
	// Need to do this for non-audio only, audio needs the artist/album directories
	CreateTopLevelDir bool              `json:"create_top_level_dir"`
	ExcludeFileRegex  string            `json:"exclude_file_regex"`
	ExtractConcurrent int               `json:"extract_concurrent"`
	TrashArchive      bool              `json:"trash_archive"`
	ExtractRename     map[string]string `json:"extract_rename"`
}

func (a *ArchiveConfig) String() string {
	return reflections.StructString(a)

}

func (a *ArchiveConfig) IsRenKey(key string) bool {
	_, ok := a.ExtractRename[key]
	return ok

}

func NewArchiveConfig() *ArchiveConfig {
	return &ArchiveConfig{
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
	}

}
