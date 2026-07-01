package platform

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed bin/*
var embeddedBinaries embed.FS

var ExtractedBinDir string

func ExtractEmbeddedBinaries() error {
	tmpDir := filepath.Join(os.TempDir(), "nexforge-bin")
	if stat, err := os.Lstat(tmpDir); err == nil && stat.Mode()&os.ModeSymlink != 0 {
		_ = os.Remove(tmpDir)
	}
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return err
	}

	entries, err := embeddedBinaries.ReadDir("bin")
	if err != nil {
		// No bin directory or it's empty
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "README.md" {
			continue
		}

		inPath := filepath.Join("bin", entry.Name())
		outPath := filepath.Join(tmpDir, entry.Name())

		// Read from embed
		data, err := embeddedBinaries.ReadFile(inPath)
		if err != nil {
			return err
		}

		// Hash comparison
		if stat, err := os.Lstat(outPath); err == nil && stat.Mode().IsRegular() {
			if f, err := os.Open(outPath); err == nil {
				h := sha256.New()
				if _, err := io.Copy(h, f); err == nil {
					f.Close()
					diskHash := h.Sum(nil)
					embedHash := sha256.Sum256(data)
					if bytes.Equal(diskHash, embedHash[:]) {
						continue // Skips disk write if already matching
					}
				} else {
					f.Close()
				}
			}
		}

		// Write to temp dir
		if err := os.WriteFile(outPath, data, 0700); err != nil {
			return err
		}
	}

	ExtractedBinDir = tmpDir
	return nil
}

func GetBinaryPath(name string) string {
	if ExtractedBinDir == "" {
		return ""
	}
	binPath := filepath.Join(ExtractedBinDir, name)
	if _, err := os.Stat(binPath); err == nil {
		return binPath
	}
	return ""
}

// ResolveCommandPath safely locates the binary (either embedded, system PATH, or local).
func ResolveCommandPath(name string) (string, error) {
	if binPath := GetBinaryPath(name); binPath != "" {
		return binPath, nil
	} else if runtime.GOOS == "windows" {
		if binPath := GetBinaryPath(name + ".exe"); binPath != "" {
			return binPath, nil
		}
	}

	if _, err := exec.LookPath(name); err != nil {
		checkPath := "./" + name
		if runtime.GOOS == "windows" {
			checkPath += ".exe"
		}
		if _, err := os.Stat(checkPath); err != nil {
			return "", fmt.Errorf("%s not found in embedded binaries, PATH, or local folder", name)
		}
		return checkPath, nil
	}
	return name, nil
}
