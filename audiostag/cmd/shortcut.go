package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

//go:embed show-archive.xml
var embeddedXML []byte

type Shortcut string

const (
	showTagsSC       Shortcut = "show-tags"
	extractArchiveSC Shortcut = "extract-archive"
)

func CreateShortcut(sc Shortcut) {
	if runtime.GOOS == "darwin" {
		slog.Debug("GOOS|darwin")
		createMacShortcut(sc)

	}

}

func createMacShortcut(sc Shortcut) error {
	tempDir := os.TempDir()

	executablePath, err := executablePath()
	if err != nil {
		return err

	}

	scName := fmt.Sprintf("%s", sc)
	shortcutTempPath := filepath.Join(tempDir, scName+"-unsigned.shortcut")

	cmd := exec.Command("plutil", "-convert", "binary1", "-o", shortcutTempPath, "-")
	replacedBytes := XMLReplace(executablePath, sc)
	cmd.Stdin = bytes.NewReader(replacedBytes)
	if err := cmd.Run(); err != nil {
		slog.Error("cmd.Run()|error executing plutil", "err", err)
		return err

	}

	signedShortcutTempPath := filepath.Join(tempDir, scName+".shortcut")
	if _, err := exec.Command("shortcuts", "sign", "-i", shortcutTempPath, "-o", signedShortcutTempPath).Output(); err != nil {
		return err

	}
	os.Remove(shortcutTempPath)

	// Create the Shortcut
	exec.Command("open", signedShortcutTempPath).Output()

	defer os.Remove(signedShortcutTempPath)
	time.Sleep(1 * time.Second)

	return nil

}

// XMLReplace points the executable path to this executable and sets the flags.
//   - executablePath is the full path to this executable
//   - sc is the Shortcut to create
//
// Return the embedded XML data replaced with the proper values as a byte slice.
func XMLReplace(executablePath string, sc Shortcut) []byte {
	const spaceFlagSpaceQuote = 9
	executablePathLen := len(executablePath)
	lengthToPlaceholder := executablePathLen + spaceFlagSpaceQuote

	scName := fmt.Sprintf("%s", sc)

	oldBytes := []byte("{NAME}")
	newBytes := []byte(scName)
	nameBytes := bytes.ReplaceAll(embeddedXML, oldBytes, newBytes)

	oldBytes = []byte("{EXECUTABLE}")
	newBytes = []byte(executablePath)
	executableBytes := bytes.ReplaceAll(nameBytes, oldBytes, newBytes)

	oldBytes = []byte("{FLAGS}")
	if sc == showTagsSC {
		newBytes = []byte("--show-tags-chooser --out window")

	} else {
		newBytes = []byte("--extract")

	}

	lenToPlaceholderBytes := bytes.ReplaceAll(executableBytes, oldBytes, newBytes)

	oldBytes = []byte("{LENGTH_TO_PLACEHOLDER}")
	newBytes = make([]byte, lengthToPlaceholder)
	newBytes = []byte("{" + strconv.Itoa(lengthToPlaceholder) + ", 1}")

	return bytes.ReplaceAll(lenToPlaceholderBytes, oldBytes, newBytes)

}

func executablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		slog.Error("os.Executable()|unable to retrieve executable path", "err", err)

		return "", err

	}

	return path, nil

}
