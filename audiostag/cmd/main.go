// Package main provides functionality for the CLI.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/pflag"

	"github.com/michaeluuong/utilize/reflections"
	"github.com/michaeluuong/utilize/stringy"

	"github.com/michaeluuong/audios/audiostag/internal/a_global"
	_ "github.com/michaeluuong/audios/audiostag/internal/a_global" // Set global configs
	"github.com/michaeluuong/audios/audiostag/internal/app/core/archiver"
	"github.com/michaeluuong/audios/audiostag/internal/app/core/mp3"
	"github.com/michaeluuong/audios/audiostag/internal/app/core/mp3/config"
	"github.com/michaeluuong/audios/audiostag/internal/pkg/aud_io"
)

const lineLength = 130

var audiostagConfig, _ = a_global.FromConfigMan()

type PflagArgs struct {
	dirFlag            *string
	extractExcludeFlag *string
	fileFlag           *string
	outFlag            *string
	pathFlag           *string

	preConditionsFlag *string
	preReplaceFlag    *string
	replaceFlag       *string
	tagFlag           *string
	keepTagFlag       *string

	albumFolderNameFlag *string
	caseFlag            *string
	coverArtistFlag     *string
	coverAlbumFlag      *string
	coverFlag           *string
	coverSourceFlag     *string
	multiDiscFlag       *string
	noArtistSplit       *bool
	playlistNameFlag    *string
	singleArtistFlag    *string
	variousFlag         *string

	albumFlag           *bool
	dummyFilesFlag      *bool
	extractFlag         *bool
	listFlag            *bool
	noDirectoryRename   *bool
	playlistFlag        *bool
	printTagsFlag       *bool
	setDefaultTagsFlag  *bool
	showArchiveTagsFlag *bool
	showTagsFlag        *bool
	showTagsChooserFlag *bool

	createExtractArchiveShortcutFlag *bool
	createShowTagsShortcutFlag       *bool

	testFlag *bool
	intFlag  *int
}

func (p PflagArgs) String() string {
	return reflections.StructString(p)

}

var pArg = PflagArgs{
	dirFlag:            pflag.StringP("dir", "d", "", "directory"),
	extractExcludeFlag: pflag.String("extract-exclude", "", "exclude this file or regular expression from being extracted"),
	fileFlag:           pflag.StringP("file", "f", "", "file expression"),
	outFlag:            pflag.StringP("out", "o", "", "write output here [file, window]"),
	pathFlag:           pflag.String("path", "p", "full path to the file"),

	preConditionsFlag: pflag.String("precondition", "", "set a tag if this condition is present else use that condition [Title= (Remastered)=Disc Number=1/2=2/2]"),
	preReplaceFlag:    pflag.String("prereplace", "", "perform this replacement at the beginning of process [tag=regexp=replace~...]"),
	replaceFlag:       pflag.StringP("replace", "r", "", "replace text in existing field [field=regexp=replace~...] (e.g. album=(Deluxe)=~title=ft.=feat)"),
	tagFlag:           pflag.StringP("tag", "t", "", "swap tag content with text [tag_field=value~...] (e.g. genre=Pop~artist=Tool~APIC=path/URL)"),
	keepTagFlag:       pflag.StringP("keep-tag", "k", "", "do not remove the tag [tag_field~tag_field2~...] (e.g. TCOM)"),

	albumFolderNameFlag: pflag.String("album-folder-name", "", ""),
	caseFlag:            pflag.String("case", "", "string case of Title, Artis & Album fields [Title, Lower, Upper]"),
	coverAlbumFlag:      pflag.String("cover-album", "", "use album to search for cover art musicbrainz/coverartarchive"),
	coverArtistFlag:     pflag.String("cover-artist", "", "use artist to search for cover art in musicbrainz/coverartarchive"),
	coverFlag:           pflag.String("cover", "", "Artist|Album to get cover art and directory to place the file"),
	coverSourceFlag:     pflag.StringP("cover-source", "c", "", "URL or path to cover art file (jpg|png)"),
	multiDiscFlag:       pflag.String("multi-disc", "", "multi-disc, [true|false]"),
	noArtistSplit:       pflag.Bool("no-artist-split", false, "do not split comma separated artists"),
	playlistNameFlag:    pflag.String("playlist-name", "", "name of playlist"),
	singleArtistFlag:    pflag.String("single-artist", "", "don't split apart comma-separated artist"),
	variousFlag:         pflag.String("various", "", "various artists [true|false]"),

	albumFlag:           pflag.Bool("album", false, "extract and tag with default values"),
	dummyFilesFlag:      pflag.Bool("dummy-files", false, "copy archive into empty files"),
	extractFlag:         pflag.BoolP("extract", "e", false, "extract files from archive, requires the path to the archive"),
	listFlag:            pflag.BoolP("list", "l", false, "list the contents of archive, requires the path to the archive file"),
	noDirectoryRename:   pflag.Bool("no-directory-rename", false, "do not rename directories (i.e. artist/album)"),
	playlistFlag:        pflag.BoolP("playlist", "", false, "create an m3u playlist for all files in a directory"),
	printTagsFlag:       pflag.Bool("print-tags", false, "display tags"),
	setDefaultTagsFlag:  pflag.Bool("set-default-tags", false, "set tag(s) to default values for all mp3 files in a directory "),
	showArchiveTagsFlag: pflag.BoolP("show-archive-tags", "s", false, "display tag information contained in an archive, requires directory and file"),
	showTagsFlag:        pflag.Bool("show-tags", false, "display tag information, requires directory and file"),
	showTagsChooserFlag: pflag.Bool("show-tags-chooser", false, "display tag information for a folder or archive"),

	createExtractArchiveShortcutFlag: pflag.Bool("create-extract-archive-shortcut", false, "add a extract-archive shortcut to right-click context"),
	createShowTagsShortcutFlag:       pflag.Bool("create-show-tags-shortcut", false, "add a show-tags shortcut to right-click context"),

	testFlag: pflag.Bool("test", false, "test something"),
	intFlag:  pflag.IntP("int", "i", 0, "int message"),
}

func init() {
}

func mp3TagAttributes() *config.TagAttributes {
	return &config.TagAttributes{
		Tags:            *pArg.tagFlag,
		PreConditions:   *pArg.preConditionsFlag,
		PreReplacements: *pArg.preReplaceFlag,
		Replacements:    *pArg.replaceFlag,
		KeepTag:         *pArg.keepTagFlag,

		AlbumFolderName:   *pArg.albumFolderNameFlag,
		CoverSource:       *pArg.coverSourceFlag,
		CoverArtist:       *pArg.coverArtistFlag,
		CoverAlbum:        *pArg.coverAlbumFlag,
		IsPlaylist:        true,
		MultiDisc:         *pArg.multiDiscFlag,
		NoDirectoryRename: *pArg.noDirectoryRename,
		PlaylistName:      *pArg.playlistNameFlag,
		SingleArtist:      *pArg.singleArtistFlag,
		StringCase:        stringy.DefaultCase.ToCase(*pArg.caseFlag),
		VariousArtists:    *pArg.variousFlag,
	}

}

// export COPYFILE_DISABLE=1 to keep ._ files out of .tar

func main() {
	// Interrupt/SIGINT (Ctrl+C), SIGTERM (kill)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pflag.Parse()
	slog.Info("Parse()", "pArg", fmt.Sprintf("%v\n", pArg), "pflag.Args()", pflag.Args())

	var dir, filename, filePath, ext string
	if len(pflag.Args()) > 0 {
		if pflag.Arg(0) == "?" || strings.ToLower(pflag.Arg(0)) == "help" {
			pflag.Usage = func() {
				fmt.Fprintf(os.Stderr, "Usage of %s: [filepath]\n", os.Args[0])
				pflag.PrintDefaults()

			}
			pflag.Usage()
			os.Exit(0)

		}

		filePath = pflag.Arg(0)
		dir, filename = filepath.Split(filePath)

	}

	// Use flag values over filePath argument while preferring --pathFlag over --dir and --file
	if *pArg.pathFlag != "" {
		filePath = *pArg.pathFlag
		ext = filepath.Ext(filePath)
		if ext == "" {
			dir = filePath

		} else {
			dir, filename = filepath.Split(filePath)

		}

	} else {
		if *pArg.dirFlag != "" {
			dir = *pArg.dirFlag

		} else if audiostagConfig.CLI.StartingDirectory != "" {
			dir = audiostagConfig.CLI.StartingDirectory

		}

		if *pArg.fileFlag != "" {
			filename = *pArg.fileFlag

		}

		if !strings.Contains(filename, ".") {
			dir = filepath.Join(dir, filename)

		}

		filePath = filepath.Join(dir, filename)

	}
	slog.Info("pflag", "filePath", filePath, "dir", dir, "filename", filename, "ext", ext)

	winman := aud_io.GetWinmanInstance()

	tagAttributes := mp3TagAttributes()

	stringy.PrintLine("-", lineLength, "\n", "stderr")
	if *pArg.listFlag {
		slog.Info("--list|-l listing contents of archive", "filePath", filePath)
		err := archiver.ListArchive(ctx, filePath, *pArg.outFlag)
		if err != nil {
			slog.Error("--list|-l ListArchive()", "err", err)
			panic(err)

		}

		stringy.PrintLine("-", lineLength)

	}

	if *pArg.showTagsChooserFlag {
		slog.Info("--show-tags-chooser|show tags in directory or archive")
		*pArg.showTagsFlag = true
		if filename != "" {
			ext := filepath.Ext(filename)
			if ext == ".tar" {
				*pArg.showTagsFlag = false
				*pArg.showArchiveTagsFlag = true

			} else if ext == ".mp3" {
				*pArg.showTagsFlag = true

			}

		}

	}

	if *pArg.showTagsFlag {
		slog.Info("--show-tags|print tags in dir", "dir", dir)

		errChan := make(chan error, 1)
		var myfunc = func() {
			err := mp3.PrintTags(dir, *pArg.outFlag, "showActions")
			errChan <- err

		}

		if *pArg.outFlag == "window" {
			winman.Run(myfunc)

		} else {
			myfunc()

		}

		if err := <-errChan; err != nil {
			panic(err)

		}

		stringy.PrintLine("-", lineLength)

	} else if *pArg.showArchiveTagsFlag {
		slog.Info("--show-archive-tags|-s list tags in filePath", "filePath", filePath)

		errChan := make(chan error, 1)

		var myfunc = func() {
			//mp3.PrintTags(dir, *pArg.outFlag)
			err := mp3.PrintArchiveTags(ctx, filePath, *pArg.outFlag, "true")
			winman.InfiniteProgressStop()
			errChan <- err

		}

		if *pArg.outFlag == "window" {
			winman.InfiniteProgressStart(cancel, "audios", filePath)
			winman.Run(myfunc)

		} else {
			myfunc()

		}
		if err := <-errChan; err != nil {
			panic(err)

		}

		stringy.PrintLine("-", lineLength, "stderr")

	} else if *pArg.printTagsFlag {
		slog.Info("--print-tags|-s list tags in filePath", "filePath", filePath)

		mp3.ReadTag(filePath)
		stringy.PrintLine("-", lineLength, "stderr")

	}

	if *pArg.albumFlag {
		slog.Info("--album extracting and tagging", "filePath", filePath)

		if *pArg.extractExcludeFlag != "" {
			audiostagConfig.Archive.ExcludeFileRegex += "|" + *pArg.extractExcludeFlag

		}

		errChan := make(chan error, 1)
		var myfunc = func() {
			var extractedDirs map[string]bool
			if filename != "" {
				var err error
				extractedDirs, err = archiver.ExtractArchive(ctx, filePath)
				errChan <- err
				slog.Debug("--album ExtractArchive()", "extractedDirs", extractedDirs, "err", err)

			} else {
				slog.Debug("directory|skipping extraction", "filename", filename)
				extractedDirs = make(map[string]bool)

			}

			if len(extractedDirs) == 0 {
				extractedDirs = map[string]bool{dir: true}

			}

			if len(extractedDirs) > 0 {
				for extDir := range extractedDirs {
					fmt.Printf("setting default tags for %s\n\n", extDir)
					winman.InfiniteProgressUpdate("setting tags to default for " + extDir)
					err := mp3.SetDefaultTags(extDir, "", tagAttributes)
					errChan <- err

					if err != nil {
						slog.Error("--album SetDefaultTags()", "err", err)

					}

				}

			}

			winman.InfiniteProgressStop()

		}

		winman.InfiniteProgressStart(cancel, "audios", filePath)
		winman.SetCancel(cancel)
		winman.Run(myfunc)
		if err := <-errChan; err != nil {
			panic(err)

		}

	} else if *pArg.testFlag {
		mp3.DummyFiles(dir, filename)

	} else if *pArg.coverFlag != "" {
		slog.Info("--cover|-c attempting to find cover art", "artist|album", pArg.coverFlag, "dir", dir)

		cover := strings.Split(*pArg.coverFlag, "|")
		if len(cover) < 2 {
			panic(errors.New("--cover|-c must provide artist and album"))

		}
		artist, album := cover[0], cover[1]

		if _, err := mp3.CoverArt(artist, album, dir); err != nil {
			slog.Error("--cover|-c CoverArt()", "err", err)
			panic(err)

		}

	} else if *pArg.dummyFilesFlag {
		slog.Info("--dummy-files|creating empty files from archive", "dir", dir, "filename", filename)

		errChan := make(chan error, 1)
		var myfunc = func() {
			err := mp3.DummyFiles(dir, filename)
			winman.InfiniteProgressStop()
			errChan <- err

		}

		winman.InfiniteProgressStart(cancel, "audios", filePath)
		winman.Run(myfunc)

		if err := <-errChan; err != nil {
			panic(err)

		}

	} else if *pArg.extractFlag {
		slog.Info("--extract|-e extracting files from archive", "dir", dir, "filename", filename)

		errChan := make(chan error, 1)
		var myfunc = func() {
			err := archiver.ExtractArchives(ctx, dir, filename)
			winman.InfiniteProgressStop()
			errChan <- err

		}

		winman.InfiniteProgressStart(cancel, "audios", filePath)
		winman.Run(myfunc)

		if err := <-errChan; err != nil {
			panic(err)

		}

	} else if *pArg.playlistFlag {
		slog.Info("--playlist|-p", "dir", dir)

		disc, err := mp3.NewDisc()
		fmt.Printf("disc: %s\n", disc)
		if err != nil {
			panic(err)

		}
		disc.AddAll(dir, tagAttributes)

		playlist := mp3.NewPlaylist()
		//playlist.CreatePlaylist(dir)
		playlist.ScratchPlaylist(disc, *pArg.playlistNameFlag)

	} else if *pArg.setDefaultTagsFlag {
		slog.Info("--set-def-tag set tags to default values", "dir", dir, "filename", filename)
		err := mp3.SetDefaultTags(dir, filename, tagAttributes)

		if err != nil {
			slog.Error("--set-def-tag SetDefaultTags()", "err", err)
			panic(err)

		}

		//mp3.DownloadPicture("https://coverartarchive.org/d75edc2b-9be9-4197-add7-22d1bd4e44c9", dir)

	} else if *pArg.tagFlag != "" {
		slog.Info("--tag|-t setting tag(s)", "dir", dir, "tag=value", *pArg.tagFlag)

		mp3.SetTagsDir(dir, tagAttributes)

	} else if *pArg.createExtractArchiveShortcutFlag {
		slog.Info("--create-extract-archive-shortcut|creating extract-archive shortcut")

		CreateShortcut(extractArchiveSC)

	} else if *pArg.createShowTagsShortcutFlag {
		slog.Info("--create-show-tags-shortcut|creating show-tags shortcut")

		CreateShortcut(showTagsSC)

	}

}
