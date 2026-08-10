// Package archiver collects functions for processing archives.
package archiver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/laurent22/go-trash"
	"github.com/mholt/archives"

	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/routine"
	"github.com/michaeluuong/utilize/stringy"

	"github.com/michaeluuong/audios/audiostag/internal/a_global"
	"github.com/michaeluuong/audios/audiostag/internal/pkg/aud_io"
)

const (
	dirNameSep string = "_"
)

var audiosConfig, _ = a_global.FromConfigMan()
var archConfig = audiosConfig.Archive

// MakeNextRootdir creates dirname if the archive doesn't already create a root, CreateTopLevelDir is true,
// or there are multiple root level directories.
// If it doesn't already exist the name will be dirname otherwise an _number suffix will be appended to dirname (e.g. dirname_2).
//   - fsys is the archive file system object containing archive information
//   - dirname is the name of the directory to create
//
// Return
//   - the directory name if created or empty string if not created
//   - error if unable to create the directory
func MakeNextRootDir(fsys archives.ArchiveFS, dirname string) (string, error) {
	rootDirs := RootDirs(fsys)

	rootDirNum := len(rootDirs)
	// If any file does not have anywhere to live create a parent directory
	_, hasNoRoot := rootDirs["empty_root_dir"]
	if hasNoRoot {
		rootDirNum--

	}

	var nextRootDir string
	// We don't have a directory yet and a file has no root, we want a parent directory or there are multiple root directories in the archive
	if hasNoRoot || archConfig.CreateTopLevelDir || rootDirNum > 1 {
		var err error
		nextRootDir, err = filing.NextDir(dirname, dirNameSep)
		if err != nil {
			slog.Error("filing.NextDir()|could not create next directory", "dirname", dirname, "err", err)
			return nextRootDir, err

		}

		if err := os.MkdirAll(nextRootDir, filing.DirPerm); err != nil {
			slog.Error("os.MkdirAll()|could not make directory", "filing.DirPerm", filing.DirPerm, "err", err)
			return nextRootDir, err

		}
		slog.Info("created directory", "nextRootDir", nextRootDir, "filing.DirPerm", filing.DirPerm)

	}

	return nextRootDir, nil

}

// ExtractArchiveFile extracts files from an archive. If there are no root directories in the archive or there are multiple
// root directories in the archive a top-level directory will be created to extract the files into. IF a RAR file is detected
// within a TAR file it will be extracted.
//   - ctx is the context for canceling
//   - archiveFilename is the name of the archive file to extract
//
// Configs
//   - create_top_level_dir if true creates a root/parent directory
//   - exclude_file_regex the files included in this config will not be extracted
//   - extract_concurrent specifies the number of concurrent extractions allowed
//   - trash_archive if true the archive will be moved to trash after extraction is completed
//   - extract_rename will rename all files matching the key extension to the value name (e.g. ".jpg": "cover" -> sound.jpg -> cover.jpg)
//
// Return the set of final directories or an error if
//   - unable to open file
//   - unable to create ArchiveFS
//   - problem walking directory
//   - cannot find an open directory name or cannot create it (e.g. root directory or top-level directory)
//   - issue reading the archive or writing the file to disk
func ExtractArchiveFiles(ctx context.Context, archiveFilename string) (map[string]bool, error) {
	winman := aud_io.GetWinmanInstance()

	fmt.Printf("\n\n----------------------------------------------------------------------------------------------\n")
	dir, filename := filepath.Split(archiveFilename)
	fmt.Printf("Extracting archive %s, dir: %s, fileanme: %s\n", archiveFilename, dir, filename)
	//wd, _ := os.Getwd()
	archiveFile, err := os.Open(archiveFilename)
	if err != nil {
		return nil, err

	}
	defer archiveFile.Close()

	fsys, format, err := archivesExtractorFS(ctx, archiveFile)
	if err != nil {
		slog.Error("archivesExtractorFS()|could not create archive.archivesFS", "archiveFile", archiveFile)
		return nil, err

	}

	archiveExt := format.Extension()
	archiveFileNoExt := strings.TrimSuffix(archiveFilename, archiveExt)
	archiveFileParts := archiveFileNoExt + "|" + archiveExt

	re, err := compileRegex(archConfig.ExcludeFileRegex)
	if err != nil {
		slog.Error("compleRegex()|could not compile regex", "Config.ExcludeFileRegex", archConfig.ExcludeFileRegex, "err", err)

	}

	var rootDir string
	finalDirs := map[string]bool{}
	var rarFilename string
	dirSet := make(map[string]string)
	fs.WalkDir(fsys, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			slog.Error("fs.WalkDir()|killed")
			return nil

		default:
			name := dirEntry.Name()
			ext := filepath.Ext(name)
			if err != nil {
				slog.Error("WalkDir()|problem walking archive", "archiveFileParts", archiveFileParts, "err", err)
				return err

			} else if path == "." || dirEntry.Name() == ".DS_Store" {
				slog.Debug("WalkDir()|skipping", "archiveFileParts", archiveFileParts, "path", path)
				return nil

			}

			// Don't want files extracted all over the current directory so try and make a root directory (but not for RAR files)
			if rootDir == "" && archiveExt != ".rar" {
				rootDir, err = MakeNextRootDir(fsys, archiveFileNoExt)
				if err != nil {
					slog.Error("MakeNextRootDir()|could not make root directory", "archiveFileParts", archiveFileParts, "rootDir", rootDir, "err", err)
					return err

				}

				if rootDir == "" {
					rootDir = dir

				}

			}

			var target string

			pathParts := strings.Split(path, string(filepath.Separator))
			newPath := filepath.Join(pathParts[1:]...)
			// In-archive directory
			if dirEntry.IsDir() {
				newDir := filepath.Join(rootDir, pathParts[0])
				if path, ok := dirSet[pathParts[0]]; !ok {
					target, err = filing.NextDir(newDir)
					if err != nil {
						return err

					}
					dirSet[pathParts[0]] = target

				} else {
					target = filepath.Join(path, newPath)

				}

				slog.Debug("os.MkdirAll()|making archive dir", "rootDir", rootDir, "newDir", newDir, "target", target)
				return os.MkdirAll(target, filing.DirPerm)

			} else {
				fileDir, filename := filepath.Split(path)
				fileDir = filepath.Clean(fileDir)
				if fileDir == "." {
					target = filepath.Join(rootDir, filename)

				} else {
					target = filepath.Join(dirSet[pathParts[0]], newPath)

				}

			}

			// Rename these files
			if newName, ok := archConfig.ExtractRename[ext]; ok {
				if strRe, ok := archConfig.ExtractRename["exclude"]; ok {
					excludeRegexp := regexp.MustCompile(strRe)
					if !excludeRegexp.MatchString(name) {
						tDir, _ := filepath.Split(target)
						fullTarget := filepath.Join(tDir, newName+ext)
						target = filing.NextFile(fullTarget)
						slog.Info("renaming file", "strRe", strRe, "target", target)

					}

				} else {
					slog.Debug("ExtractRename|file excluded", "newName", newName, "strRe", strRe)

				}

			}

			// Skip these files (but not directories)
			if re != nil && re.MatchString(name) {
				slog.Debug("re.MatchString()|skip file", "archiveFileParts", archiveFileParts, "name", name, "ext", ext, "ExcludeFileRegex", archConfig.ExcludeFileRegex)
				return nil

			}

			if filepath.Ext(name) == ".rar" {
				rarFilename = target
				slog.Debug("filepath.Ext()|found rar file", "archiveFileParts", archiveFileParts, "rarFilename", rarFilename)

			}

			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				slog.Error("fs.ReadFile()|could not read file", "archiveFileParts", archiveFileParts, "path", path, "err", err)
				return err

			}

			finalDir, _ := filepath.Split(target)
			finalDirs[finalDir] = true

			slog.Info("extracting file", "target", target, "permissions", filing.FilePerm)
			winman.InfiniteProgressUpdate("extracting " + target)
			return os.WriteFile(target, data, filing.FilePerm)
		}

	})

	if rarFilename != "" {
		rarDir, rarFile := filepath.Split(rarFilename)
		return finalDirs, ExtractRars(ctx, filepath.Clean(rarDir), rarFile)

	}

	return finalDirs, nil

}

// ExtractRars extracts from a multi-part (split) RAR file (all files must be in the same directory).
//   - ctx is the context used by mholt's archives.Extract for cancellation
//   - filename is the full path to the split RAR with the .rar extension
//
// Return an error
//   - can't make directories
//   - can't open rar file
//   - can't create the output file
//   - can't copy data
//
// func (e Extracter) ExtractRars(ctx context.Context, dir, rarFilename string) error {
func ExtractRars(ctx context.Context, dir, filename string) error {
	if dir == "" || filename == "" {
		slog.Error("parameter(s) must not be empty", "dir", dir, "filename", filename)
		return errors.New("parameter(s) must not be empty, dir: " + dir + ", filename: " + filename)

	}

	rar := archives.Rar{
		Name: filename,
		FS:   archives.DirFS(dir),
	}

	return rar.Extract(ctx, nil, func(ctx context.Context, file archives.FileInfo) error {
		if strings.Contains(file.NameInArchive, ".DS_Store") {
			return nil

		}

		outPath := filepath.Join(dir, file.NameInArchive)
		if file.IsDir() {
			return os.MkdirAll(outPath, filing.DirPerm)

		}

		if err := os.MkdirAll(dir, filing.DirPerm); err != nil {
			return err

		}

		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		out, err := os.Create(outPath)
		if err != nil {
			return err

		}
		defer out.Close()

		_, err = io.Copy(out, rc)
		return err

	})

}

func extractChooser(ctx context.Context, filePath string) (map[string]bool, error) {
	emptyMap := map[string]bool{}

	dir, filename := filepath.Split(filePath)
	format, err := IdentifyFormat(ctx, filePath)
	if err != nil {
		return emptyMap, err

	}

	var archiveDirs map[string]bool
	if format != nil && format.Extension() == ".rar" {
		err = ExtractRars(ctx, dir, filename)
		if err != nil {
			//errChan <- err
			return emptyMap, err

		}

	} else {
		archiveDirs, err = ExtractArchiveFiles(ctx, filePath)
		if err != nil {
			slog.Error("ExtractArchiveFiles()|could not extract archive", "filePath", filePath, "err", err)
			return archiveDirs, err

		}

	}

	if archConfig.TrashArchive {
		if _, err := trash.MoveToTrash(filePath); err != nil {
			slog.Error("MoveToTrash()|could not move to trash", "TrashArchive", archConfig.TrashArchive, "filePath", filePath, "err", err)
			// errChan <- err
			return archiveDirs, err

		}

	}

	return archiveDirs, nil

}

// ExtractArchive extracts archive files one at a time. See ExtractArchives for concurrent extractions.
//   - ctx is the context used for canceling
//   - filePath is the path to the file(s) to unarchive
//
// Return a set of archives that were extracted.
func ExtractArchive(ctx context.Context, filePath string) (map[string]bool, error) {
	emptyMap := map[string]bool{}
	if filePath == "" {
		return emptyMap, errors.New("must provide a filePath")
	}

	dir, _ := filepath.Split(filePath)

	owd, err := os.Getwd()
	if err != nil {
		return emptyMap, err

	}

	if err := os.Chdir(dir); err != nil {
		return emptyMap, err

	}

	if !filing.Exists(filePath) {
		return emptyMap, errors.New("file does not exist: " + filePath)

	}

	archiveDirs, err := extractChooser(ctx, filePath)
	if err != nil {
		return emptyMap, nil

	}

	return archiveDirs, os.Chdir(owd)

}

// ExtractArchives
func ExtractArchives(ctx context.Context, dir string, filenameRegexOpt ...string) error {
	filenameRegex := ""
	if len(filenameRegexOpt) > 0 {
		filenameRegex = regexp.QuoteMeta(filenameRegexOpt[0])

	}

	owd, err := os.Getwd()
	if err != nil {
		return err

	}

	if err := os.Chdir(dir); err != nil {
		return err

	}

	archiveDirList := filing.LsEntryName(dir, filenameRegex, "add_dir")
	extractConcurrent := archConfig.ExtractConcurrent
	sem := routine.NewSemaphore(extractConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	const runGroup = 1
	for _, archivePath := range archiveDirList {
		wg.Add(runGroup)
		go func() {
			defer wg.Done()
			err := sem.RunContext(ctx, func() error {
				slog.Debug("RunContext()|archiveFileInfo", "archivePath", archivePath)
				_, err := extractChooser(ctx, archivePath)
				return err

			})
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()

			}

		}()

	}
	wg.Wait()

	return os.Chdir(owd)

}

// streamSize determines the size of the stream.
//   - s is the stream to find the size of
//
// Return
//   - size of stream s as an int64
//   - an error if unable to seek at any point in determining the size of the stream
func streamSize(s io.Seeker) (int64, error) {
	cur, err := s.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err

	}

	end, err := s.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err

	}

	_, err = s.Seek(cur, io.SeekStart)
	if err != nil {
		return 0, err

	}

	return end, nil

}

func ListArchive(ctx context.Context, archiveFilename string, outFilenameOpt ...string) error {
	archiveFile, err := os.Open(archiveFilename)
	if err != nil {
		return err

	}
	defer archiveFile.Close()

	var writer io.Writer = os.Stdout
	if len(outFilenameOpt) > 0 && outFilenameOpt[0] != "" {
		oFile, err := os.OpenFile(outFilenameOpt[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, filing.FilePerm)
		if err != nil {
			slog.Error("os.OpenFile()|could not open file", "outFilename", outFilenameOpt[0])

		}
		defer oFile.Close()
		writer = oFile

	}

	rowNum := 0
	data := make(map[int][]string)
	header := []string{"Archive", "IsDir", "Mode", "Size", "ModTime", "File"}
	data[rowNum] = header
	rowNum++

	fsys, _, err := archivesExtractorFS(ctx, archiveFile)
	fs.WalkDir(fsys, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("WalkDir()|could not walk directory", "dirEntry.Name()", dirEntry.Name(), "err", err)
			return err

		} else if path == "." {
			return nil

		}
		info, err := dirEntry.Info()
		row := []string{
			filepath.Base(archiveFilename),
			fmt.Sprintf("%t", info.IsDir()),
			fmt.Sprintf("%s", info.Mode()),
			fmt.Sprintf("%d", info.Size()),
			fmt.Sprintf("%s", info.ModTime()),
			path,
		}
		data[rowNum] = row
		rowNum++

		return nil

	})
	stringy.PrintData("", data, writer)

	return nil

}

// RootDirs looks for root-level directories in the archive.
// If any files do not have a parent directory to live in, a key of "empty_root_dir" will be placed in the set.
//   - fys contains the walkable archive information
//
// Return the set of root-level directories
func RootDirs(fsys archives.ArchiveFS) map[string]bool {
	rootDirs := map[string]bool{}
	fs.WalkDir(fsys, ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err

		}

		if dirEntry.IsDir() || strings.Contains(path, filing.PathSep) {
			dirnames := strings.Split(path, filing.PathSep)
			if dirnames[0] != "" && dirnames[0] != "." {
				rootDirs[dirnames[0]] = true

			}

		} else {
			rootDirs["empty_root_dir"] = true

		}

		return nil

	})
	slog.Debug("WalkDir()|found roots", "rootDirs", rootDirs)

	return rootDirs

}

func ArchivesExtractorFS(ctx context.Context, archiveFile *os.File) (archives.ArchiveFS, archives.Format, error) {
	return archivesExtractorFS(ctx, archiveFile)

}

func archivesExtractorFS(ctx context.Context, archiveFile *os.File) (archives.ArchiveFS, archives.Format, error) {
	emptyFsys := archives.ArchiveFS{}

	format, stream, err := archives.Identify(ctx, archiveFile.Name(), archiveFile)
	slog.Info("archive format identified", "format.Extension", format.Extension(), "format.MediaType()", format.MediaType())
	if err != nil {
		return emptyFsys, format, err

	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		slog.Error("format()|failed to assert format to archives.Extractor", "ok", ok)
		return emptyFsys, format, errors.New("Could not assert archives.Format to archives.Extractor")

	}

	ras, ok := stream.(archives.ReaderAtSeeker)
	if !ok {
		slog.Error("stream()|failed to assert stream to archives.ReadAtSeeker", "archiveFile", archiveFile, "err", err)
		return emptyFsys, format, errors.New("stream must implement io.ReaderAt and io.Seeker for ArchiveFS")

	}

	size, err := streamSize(ras)
	if err != nil {
		return emptyFsys, format, err

	}

	sectionReader := io.NewSectionReader(ras, 0, size)

	fsys := archives.ArchiveFS{
		Format:  extractor,
		Stream:  sectionReader,
		Context: ctx,
	}

	return fsys, format, err

}

// func (e Extracter) Identify(ctx context.Context, dirstring, filenameRegexOpt ...string) error {
func IdentifyFormat(ctx context.Context, filename string) (archives.Format, error) {
	file, err := os.Open(filename)
	if err != nil {
		slog.Error("os.Open|could not open file", "filename", filename, "err", err)
		return nil, err

	}
	defer file.Close()

	format, _, err := archives.Identify(ctx, filename, file)
	if err != nil {
		slog.Error("archives.Identify()|could not identify archive", "filename", filename, "err", err)
		return nil, err

	}
	slog.Debug("archives.Identify()|identfied format", "filename", filename, "Extension", format.Extension(), "MediaType", format.MediaType())

	switch f := format.(type) {
	case archives.CompressedArchive:
		slog.Debug("archives.Identify()|identified type", "type", "archives.CompressedArchive", "f", f)

	case archives.Compression:
		slog.Debug("archives.Identify()|identified type", "type", "archives.Compression", "f", f)

	case archives.Archival:
		slog.Debug("archives.Identify()|identified type", "type", "archives.Archival", "f", f)

	case archives.Extraction:
		slog.Debug("archives.Identify()|identified type", "type", "archives.Extraction", "f", f)

	}

	return format, nil
}

// compileRegex returns the result of compiling the regular expression parameter.
//   - regex is the regular expression to compile
//
// Return
//   - re is the compiled regex expression
//   - err is the error if regex could not be compiled
func compileRegex(regex string) (re *regexp.Regexp, err error) {
	if regex != "" {
		re, err = regexp.Compile(regex)
		if err != nil {
			re = nil
			slog.Error("regexp.Compile()|could not compile regex", "regex", regex)

		}

	}

	return

}
