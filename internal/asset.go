package internal

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	modelVersion     = "v1.0.0"
	modelFilename    = "kokoro-v0_19.onnx"
	modelURL         = "https://github.com/nazdridoy/kokoro-tts/releases/download/v1.0.0/kokoro-v0_19.onnx"
	voicesVersion    = "v1.0.0"
	voicesArchive    = "voices-v1.0.bin"
	voicesURL        = "https://github.com/nazdridoy/kokoro-tts/releases/download/v1.0.0/voices-v1.0.bin"
	runtimeVersion   = "1.25.1"
)

var (
	runtimeURLs = map[string]string{
		"linux":   fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/v%s/onnxruntime-linux-x64-%s.tgz", runtimeVersion, runtimeVersion),
		"darwin":  fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/v%s/onnxruntime-osx-universal2-%s.tgz", runtimeVersion, runtimeVersion),
		"windows": fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/v%s/onnxruntime-win-x64-%s.zip", runtimeVersion, runtimeVersion),
	}
	runtimeLibNames = map[string]string{
		"linux":   "libonnxruntime.so",
		"darwin":  "libonnxruntime.dylib",
		"windows": "onnxruntime.dll",
	}
)

func AssetDir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		switch runtime.GOOS {
		case "darwin":
			base = filepath.Join(home, "Library", "Application Support")
		case "windows":
			base = os.Getenv("APPDATA")
			if base == "" {
				base = filepath.Join(home, "AppData", "Roaming")
			}
		default:
			base = filepath.Join(home, ".local", "share")
		}
	}
	assetDir := filepath.Join(base, "vox", "assets")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		return "", err
	}
	return assetDir, nil
}

func ModelPath() (string, error) {
	dir, err := AssetDir()
	if err != nil {
		return "", err
	}
	modelDir := filepath.Join(dir, "models")
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(modelDir, modelFilename), nil
}

func VoiceDir() (string, error) {
	dir, err := AssetDir()
	if err != nil {
		return "", err
	}
	voiceDir := filepath.Join(dir, "voices")
	if err := os.MkdirAll(voiceDir, 0o755); err != nil {
		return "", err
	}
	return voiceDir, nil
}

func RuntimeLibPath() (string, error) {
	dir, err := AssetDir()
	if err != nil {
		return "", err
	}
	runtimeDir := filepath.Join(dir, "runtime")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return "", err
	}
	libName := runtimeLibNames[runtime.GOOS]
	if libName == "" {
		libName = "libonnxruntime.so"
	}
	return filepath.Join(runtimeDir, libName), nil
}

func EnsureModel() (string, error) {
	path, err := ModelPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	fmt.Fprintf(os.Stderr, "Downloading Kokoro model (%s)...\n", modelVersion)
	if err := downloadWithProgress(modelURL, path); err != nil {
		return "", fmt.Errorf("failed to download model: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Model downloaded to %s\n", path)
	return path, nil
}

func EnsureVoices() (string, error) {
	voiceDir, err := VoiceDir()
	if err != nil {
		return "", err
	}

	afHeartPath := filepath.Join(voiceDir, "af_heart.bin")
	if _, err := os.Stat(afHeartPath); err == nil {
		return voiceDir, nil
	}

	fmt.Fprintf(os.Stderr, "Downloading voice profiles (%s)...\n", voicesVersion)
	tmpArchive, err := os.CreateTemp("", "voices-*.bin")
	if err != nil {
		return "", err
	}
	tmpArchive.Close()
	defer os.Remove(tmpArchive.Name())

	if err := downloadWithProgress(voicesURL, tmpArchive.Name()); err != nil {
		return "", fmt.Errorf("failed to download voices: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Extracting voice profiles...\n")
	if err := extractVoices(tmpArchive.Name(), voiceDir); err != nil {
		return "", fmt.Errorf("failed to extract voices: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Voice profiles extracted to %s\n", voiceDir)
	return voiceDir, nil
}

func EnsureRuntime() (string, error) {
	path, err := RuntimeLibPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	url := runtimeURLs[runtime.GOOS]
	if url == "" {
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	fmt.Fprintf(os.Stderr, "Downloading ONNX Runtime %s for %s...\n", runtimeVersion, runtime.GOOS)
	tmpArchive, err := os.CreateTemp("", "onnxruntime-*.archive")
	if err != nil {
		return "", err
	}
	tmpArchive.Close()
	defer os.Remove(tmpArchive.Name())

	if err := downloadWithProgress(url, tmpArchive.Name()); err != nil {
		return "", fmt.Errorf("failed to download ONNX runtime: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Extracting ONNX Runtime...\n")
	if err := extractRuntime(tmpArchive.Name(), filepath.Dir(path), runtime.GOOS); err != nil {
		return "", fmt.Errorf("failed to extract ONNX runtime: %w", err)
	}

	fmt.Fprintf(os.Stderr, "ONNX Runtime extracted to %s\n", filepath.Dir(path))
	return path, nil
}

func downloadWithProgress(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "▶",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	return err
}

func extractVoices(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, f := range reader.File {
		if strings.HasSuffix(f.Name, ".npy") {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}

			name := strings.TrimSuffix(f.Name, ".npy") + ".bin"
			outPath := filepath.Join(destDir, name)

			if err := convertNPYToBin(data, outPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func convertNPYToBin(npyData []byte, outPath string) error {
	if len(npyData) < 10 {
		return fmt.Errorf("NPY file too small")
	}

	magic := []byte{0x93, 'N', 'U', 'M', 'P', 'Y'}
	if !bytesHasPrefix(npyData, magic) {
		return fmt.Errorf("invalid NPY magic header")
	}

	majorVersion := npyData[6]
	var headerLen int
	dataOffset := 0

	if majorVersion == 1 {
		if len(npyData) < 10 {
			return fmt.Errorf("NPY v1 header too small")
		}
		headerLen = int(npyData[8]) | int(npyData[9])<<8
		dataOffset = 10 + headerLen
	} else if majorVersion == 2 {
		if len(npyData) < 12 {
			return fmt.Errorf("NPY v2 header too small")
		}
		headerLen = int(npyData[8]) | int(npyData[9])<<8 | int(npyData[10])<<16 | int(npyData[11])<<24
		dataOffset = 12 + headerLen
	} else {
		return fmt.Errorf("unsupported NPY version: %d", majorVersion)
	}

	if dataOffset >= len(npyData) {
		return fmt.Errorf("NPY header exceeds file size")
	}

	rawData := npyData[dataOffset:]
	if len(rawData)%4 != 0 {
		return fmt.Errorf("NPY data size not multiple of 4")
	}

	return os.WriteFile(outPath, rawData, 0o644)
}

func bytesHasPrefix(data, prefix []byte) bool {
	if len(data) < len(prefix) {
		return false
	}
	for i := range prefix {
		if data[i] != prefix[i] {
			return false
		}
	}
	return true
}

func extractRuntime(archivePath, destDir, goos string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, destDir, goos)
	}
	return extractTgz(archivePath, destDir, goos)
}

func extractTgz(archivePath, destDir, goos string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	libName := runtimeLibNames[goos]

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Skip symlinks (typeflag 'L' or '2')
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == '2' {
			continue
		}

		baseName := filepath.Base(header.Name)
		// Look for the actual library file (with version), not symlinks
		if baseName == libName+".1.25.1" || baseName == libName+".1" || baseName == libName {
			outPath := filepath.Join(destDir, libName)
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			if goos != "windows" {
				os.Chmod(outPath, 0o755)
			}
			return nil
		}
	}
	return fmt.Errorf("library %s not found in archive", libName)
}

func extractZip(archivePath, destDir, goos string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	libName := runtimeLibNames[goos]

	for _, f := range reader.File {
		if filepath.Base(f.Name) == libName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			outPath := filepath.Join(destDir, libName)
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			return err
		}
	}
	return fmt.Errorf("library %s not found in archive", libName)
}