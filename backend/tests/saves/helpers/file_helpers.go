package helpers

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
)

// FileSize represents common file sizes for testing
type FileSize int

const (
	SizeSmall  FileSize = 1024        // 1KB
	SizeMedium FileSize = 1024 * 100  // 100KB
	SizeLarge  FileSize = 1024 * 1024 // 1MB
	SizeXLarge FileSize = 1024 * 1024 * 10 // 10MB
)

// SaveFileType represents the type of save file to generate
type SaveFileType string

const (
	TypeZip SaveFileType = "zip"
	TypeSav SaveFileType = "sav"
)

// CreateDummySaveFile creates a dummy save file of specified type and size
func CreateDummySaveFile(filePath string, fileType SaveFileType, size FileSize) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	switch fileType {
	case TypeZip:
		return createDummyZipFile(filePath, int64(size))
	case TypeSav:
		return createDummySavFile(filePath, int64(size))
	default:
		return fmt.Errorf("unsupported file type: %s", fileType)
	}
}

// createDummyZipFile creates a ZIP file with random content
func createDummyZipFile(filePath string, targetSize int64) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Calculate how many files and their sizes to reach target
	numFiles := gofakeit.Number(1, 5)
	filesCreated := 0
	remainingSize := targetSize

	for filesCreated < numFiles && remainingSize > 0 {
		fileName := fmt.Sprintf("savefile_%d.dat", filesCreated)
		fileSize := remainingSize / int64(numFiles-filesCreated)
		
		// Ensure minimum file size of 100 bytes
		if fileSize < 100 {
			fileSize = remainingSize
		}

		if err := addFileToZip(zipWriter, fileName, fileSize); err != nil {
			return err
		}

		remainingSize -= fileSize
		filesCreated++
	}

	// Add a manifest file
	manifestContent := generateGameManifest()
	manifestWriter, err := zipWriter.Create("manifest.json")
	if err != nil {
		return err
	}
	
	_, err = manifestWriter.Write([]byte(manifestContent))
	return err
}

// addFileToZip adds a file with random content to the ZIP archive
func addFileToZip(zipWriter *zip.Writer, fileName string, size int64) error {
	writer, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	// Generate random content
	randomData := make([]byte, size)
	if _, err := rand.Read(randomData); err != nil {
		return err
	}

	_, err = writer.Write(randomData)
	return err
}

// createDummySavFile creates a .sav file with game-like binary content
func createDummySavFile(filePath string, targetSize int64) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write a simple header
	header := generateSavFileHeader()
	if _, err := file.Write(header); err != nil {
		return err
	}

	// Fill the rest with random data to reach target size
	remainingSize := targetSize - int64(len(header))
	if remainingSize > 0 {
		randomData := make([]byte, remainingSize)
		if _, err := rand.Read(randomData); err != nil {
			return err
		}
		
		_, err = file.Write(randomData)
	}

	return err
}

// generateSavFileHeader creates a realistic save file header
func generateSavFileHeader() []byte {
	var buffer bytes.Buffer
	
	// Magic number
	buffer.WriteString("GAMESAV1")
	
	// Version
	buffer.WriteByte(0x01)
	buffer.WriteByte(0x00)
	
	// Player name (32 bytes, null-terminated)
	playerName := gofakeit.Gamertag()
	if len(playerName) > 31 {
		playerName = playerName[:31]
	}
	playerBytes := make([]byte, 32)
	copy(playerBytes, playerName)
	buffer.Write(playerBytes)
	
	// Game progress (4 bytes)
	progress := gofakeit.Uint32()
	buffer.WriteByte(byte(progress))
	buffer.WriteByte(byte(progress >> 8))
	buffer.WriteByte(byte(progress >> 16))
	buffer.WriteByte(byte(progress >> 24))
	
	// Timestamp (8 bytes)
	timestamp := gofakeit.Date().Unix()
	for i := 0; i < 8; i++ {
		buffer.WriteByte(byte(timestamp >> (i * 8)))
	}
	
	return buffer.Bytes()
}

// generateGameManifest creates a JSON manifest for ZIP save files
func generateGameManifest() string {
	return fmt.Sprintf(`{
	"version": "1.0",
	"game": "%s",
	"player": "%s",
	"save_date": "%s",
	"level": %d,
	"checksum": "%s"
}`, 
		gofakeit.Gamertag()+" Adventure",
		gofakeit.Gamertag(), 
		gofakeit.Date().Format("2006-01-02T15:04:05Z"),
		gofakeit.Number(1, 100),
		gofakeit.LetterN(16))
}

// CreateRandomSaveFile creates a save file with random type and size
func CreateRandomSaveFile(basePath string) (string, error) {
	// Random file type
	fileTypes := []SaveFileType{TypeZip, TypeSav}
	fileType := fileTypes[gofakeit.Number(0, len(fileTypes)-1)]
	
	// Random size
	sizes := []FileSize{SizeSmall, SizeMedium, SizeLarge}
	size := sizes[gofakeit.Number(0, len(sizes)-1)]
	
	// Generate filename
	fileName := fmt.Sprintf("%s_%s.%s", 
		gofakeit.Gamertag(),
		gofakeit.LetterN(8),
		string(fileType))
	fileName = strings.ReplaceAll(fileName, " ", "_")
	
	filePath := filepath.Join(basePath, fileName)
	err := CreateDummySaveFile(filePath, fileType, size)
	
	return filePath, err
}

// CreateTestSaveDirectory creates a temporary directory for test save files
func CreateTestSaveDirectory() (string, error) {
	return os.MkdirTemp("", "test_saves_*")
}

// CleanupTestSaveDirectory removes the test save directory and all files
func CleanupTestSaveDirectory(dirPath string) error {
	return os.RemoveAll(dirPath)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// ValidateZipFile checks if a ZIP file is valid and readable
func ValidateZipFile(filePath string) error {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	
	// Try to read each file in the ZIP
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		
		// Read a small chunk to verify readability
		buffer := make([]byte, 1024)
		_, err = io.ReadAtLeast(rc, buffer, min(1024, int(file.UncompressedSize64)))
		rc.Close()
		
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}
	}
	
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
