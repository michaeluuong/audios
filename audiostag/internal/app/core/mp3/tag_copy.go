package mp3

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"

	"github.com/michaeluuong/audios/audiostag/internal/app/core/archiver"
	"github.com/michaeluuong/audios/audiostag/internal/pkg/aud_io"

	"github.com/michaeluuong/utilize/filing"
)

var winman = aud_io.GetWinmanInstance()

func DummyFiles(dir, filename string) error {
	if !isFfmpeg() {
		return errors.New("can't find ffmpeg on the system")

	}

	srcPath := filepath.Join(dir, filename)
	srcExt := filepath.Ext(filename)
	dstDir := dir

	winman.InfiniteProgressUpdate("creating dummy files " + srcPath)

	if srcExt == "" {
		return ReadFolder(srcPath, dstDir)

	} else if srcExt == ".tar" || srcExt == ".gz" {
		return ReadArchive(srcPath, dstDir)

	}

	return nil

}

func ReadFolder(srcPath, dstDir string) error {
	if srcPath == "" || dstDir == "" {
		slog.Error("input|invalid values", "srcPath", srcPath, "dstPath", dstDir)
	}

	winman.InfiniteProgressUpdate("reading folder " + srcPath)

	srcBase := filepath.Base(filepath.Clean(srcPath))
	dstPath := filepath.Join(dstDir, srcBase) + "-audios"
	fmt.Printf("ReadFolder()|srcBase: %s, dstPath: %s\n", srcBase, dstPath)
	files := filing.LsEntryName(srcPath, "add_dir")
	for _, file := range files {
		fmt.Printf("ReadFolder()|file: %s\n", file)
	}

	return nil

}

func ReadArchive(srcPath, dstDir string) error {
	winman.InfiniteProgressUpdate("reading archive " + srcPath)

	srcDir, srcFilename := filepath.Split(filepath.Clean(srcPath))
	ext := filepath.Ext(srcFilename)
	dstTarFile := strings.TrimSuffix(srcFilename, ext) + "-audios" + ext
	dstTarFilePath := filing.NextFile(filepath.Join(srcDir, dstTarFile))
	slog.Debug("MWCHECK", "dstTarFile", dstTarFile, "dstTarFilePath", dstTarFilePath)

	file, err := os.Open(srcPath)
	if err != nil {
		return err

	}
	defer file.Close()

	winman.InfiniteProgressUpdate("extension " + ext)
	var reader io.Reader
	if ext == ".gz" {
		reader, err = gzip.NewReader(file)
		if err != nil {
			slog.Error("gzip.NewReader()|could not get reader for gzip file", "srcPath", srcPath, "err", err)
			return err
		}

	} else if ext == ".tar" {
		reader = file

	} else {
		slog.Error("only tar and gzip files are currently supported", "ext", ext)
		return errors.New("only tar and gzip files are currently supported, ext: " + ext)

	}

	tarReader := tar.NewReader(reader)
	var dstFileDir string
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				winman.InfiniteProgressUpdate("finished reading archive")

				if err := archiver.TarDirectory(context.TODO(), dstFileDir, dstTarFilePath); err != nil {
					slog.Error("TarDirectory()|tar error", "dstFileDir", dstFileDir, "dstTarFile", dstTarFile)
					return err

				}

				break

			} else {
				return err

			}

		}

		if header.FileInfo().IsDir() || header.Typeflag != tar.TypeReg {
			continue

		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tarReader); err != nil {
			slog.Error("io.Copy()|error copying to buffer continuing to next file", "err", err)
			continue

		}

		seeker := bytes.NewReader(buf.Bytes())
		dir, filename := filepath.Split(filepath.Clean(header.Name))
		if dstFileDir == "" {
			if dir == "" || strings.HasPrefix(dir, "..") || strings.HasPrefix(dir, "/") {
				srcDir = strings.TrimSuffix(srcPath, ext)

			}
			dstFileDir, _ = filing.NextDir(filepath.Join(dstDir, dir))
			os.MkdirAll(dstFileDir, filing.DirPerm)

		}

		if !strings.HasSuffix(strings.ToLower(header.Name), ".mp3") {
			dstFilePath := filepath.Join(dstFileDir, filename)
			slog.Debug("MANI", "dstFilePath", dstFilePath, "dstFileDir", dstFileDir)
			if header.Typeflag == tar.TypeReg {
				slog.Debug("MANI", "Typeflag", header.Typeflag)
				if err := os.MkdirAll(dstFileDir, filing.DirPerm); err != nil {
					slog.Error("os.MkdirAll()|error making directories", "dstFileDir", dstFileDir)
					return fmt.Errorf("failed to create parent directory: %w", err)

				}

				outFile, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
				if err != nil {
					slog.Error("os.OpenFile()|error creating file", "dstFilePath", dstFilePath)
					return fmt.Errorf("failed to create file: %w", err)

				}

				_, err = io.Copy(outFile, &buf)
				outFile.Close() // Close explicitly immediately after processing
				if err != nil {
					slog.Error("io.Copy()|error creating file", "outFile", outFile, "err", err)
					return fmt.Errorf("failed to write file content: %w", err)

				}

			}

			continue

		}

		copyTags(seeker, header, dstFileDir)

	}

	return nil

}

func copyTags(r io.ReadSeeker, header *tar.Header, dstFileDir string) error {
	_, srcFilename := filepath.Split(header.Name)
	mp3, err := tag.ReadFrom(r)
	if err != nil {
		fmt.Printf("error reading tags: %v\n", err)
		return err

	}

	dstFilePath := filepath.Join(dstFileDir, srcFilename)
	winman.InfiniteProgressUpdate("creating dummy file " + dstFilePath)
	if err := CreateDummyMP3(dstFilePath); err != nil {
		slog.Error("CreateDummyMP3()|unable to create dummy mp3", "err", err)
		return err

	}

	track, err := NewTrack(dstFilePath)
	if err != nil {
		return err

	}

	//dstFilePath := filepath.Join(dst, srcFilename)
	raw := mp3.Raw()
	for rID, rValue := range raw {
		winman.InfiniteProgressUpdate("copying tags to " + dstFilePath)
		value1, value2, _ := GetTagValues(rID, rValue)
		if rID == "APIC" {
			pictureData := rValue.(*tag.Picture).Data
			AddAPICData(track.mp3, pictureData)

		} else {
			if value2 != "" {
				track.SetTag(rID, value1, value2)

			} else {
				track.SetTag(rID, value1)

			}

		}

	}
	track.Write()

	fmt.Printf("\n")

	return nil

}

func isFfmpeg() bool {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		slog.Error("can't find ffmpeg", "err", err)

	}

	isFfmpeg := false
	if ffmpegPath != "" {
		isFfmpeg = true

	}
	return isFfmpeg

}

func CreateDummyMP3(dstPath string) error {
	args := []string{
		"-f", "lavfi",
		"-i", "anullsrc=r=44100:cl=stereo",
		"-t", "1",
		"-q:a", "9",
		"-acodec", "libmp3lame",
		dstPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return errors.New("could not create dummy mp3 file")

	}

	return nil

}
