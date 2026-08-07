// * Helps with archive creation
package archiver

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mholt/archives"
)

// TarDirectory tars up a directory.
//   - srcDir is the full path to the directory to archive
//   - tarFilePath is the full path to the name of the tar file
//
// Return an error if
//   - input is empty
//   - unable to create archive file on disk
//   - error mapping files on disk
//   - error creating tar file
func TarDirectory(ctx context.Context, srcDir, tarFilePath string) error {
	if srcDir == "" || tarFilePath == "" {
		slog.Error("input|invalid values", "srcDir", srcDir, "tarFilePath", tarFilePath)
		return fmt.Errorf("srcDir: %s and tarFilePath: %s must have valid values", srcDir, tarFilePath)

	}

	srcBase := filepath.Base(filepath.Clean(srcDir))

	return Tar(ctx, map[string]string{srcDir: srcBase}, tarFilePath)

}

// Tar archives files and directories to a single tar file.
//   - srcFileToPatha is keyed by the full path to the file/folder being tarred and valued by the tar archive name
//   - tarFilePath is the full path to the archive file being written
//
// Return an error if
//   - input is empty
//   - unable to create archive file on disk
//   - error mapping files on disk
//   - error creating tar file
func Tar(ctx context.Context, srcFileToPath map[string]string, tarFilePath string) error {
	if len(srcFileToPath) == 0 {
		slog.Error("input|invalid value", "srcFileToPath", srcFileToPath)
		return fmt.Errorf("no files to process, srcFileToPath: %s", srcFileToPath)

	} else if tarFilePath == "" {
		slog.Error("input|invalid value", "tarFilePath", tarFilePath)
		return fmt.Errorf("no target file path, tarFilePath: %s", tarFilePath)

	}

	dstFile, err := os.Create(tarFilePath)
	if err != nil {
		slog.Error("os.Create()|error creating file", "tarFilePath", tarFilePath, "err", err)
		return err

	}

	srcFiles, err := archives.FilesFromDisk(ctx, nil, srcFileToPath)
	if err != nil {
		slog.Error("FilesFromDisc()|error mapping files on disk", "srcFileToPath", srcFileToPath, "err", err)
		return err
	}

	format := archives.Tar{}

	if err := format.Archive(ctx, dstFile, srcFiles); err != nil {
		slog.Error("Archive()|error createing tar file", "dstFile", dstFile, "srcFiles", srcFiles)
		return err

	}

	return nil

}

// TarDir is the go implementation of tarring a directory.
//   - srcDir is the directory to tar up
//   - dstTarFile is the name of the resulting tar file
//
// Return an error if
//   - input is empty
//   - unable to create dstTarFile on disk
//   - unable to write tar file
func TarDir(srcDir string, dstTarFile string) error {
	if srcDir == "" || dstTarFile == "" {
		slog.Error("input|invalid values", "srcDir", srcDir, "dstTarFile", dstTarFile)
		return errors.New("srcDir: " + srcDir + " and dstTarFile: " + dstTarFile + " must have valid values")

	}

	tarFile, err := os.Create(dstTarFile)
	if err != nil {
		slog.Error("os.Create()|error creating tar file", "dstTarFile", dstTarFile)
		return err

	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	fsys := os.DirFS(srcDir)
	if err = tarWriter.AddFS(fsys); err != nil {
		slog.Error("AddFS()|error writing tar file", "srcDir", srcDir, "dstTarFile", dstTarFile)
		return err

	}

	return nil

}
