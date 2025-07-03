package saves_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"panzerstadt/async-multiplayer/tests/saves/helpers"
)

func TestCreateDummyZipFile(t *testing.T) {
	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)
	defer helpers.CleanupTestSaveDirectory(tempDir)

	filePath := filepath.Join(tempDir, "test.zip")
	err = helpers.CreateDummySaveFile(filePath, helpers.TypeZip, helpers.SizeMedium)
	require.NoError(t, err)

	// Verify file exists and has reasonable size
	size, err := helpers.GetFileSize(filePath)
	require.NoError(t, err)
	assert.Greater(t, size, int64(1000)) // Should be at least 1KB
	assert.Less(t, size, int64(200000))  // Should be less than 200KB (compressed)

	// Verify it's a valid ZIP file
	err = helpers.ValidateZipFile(filePath)
	assert.NoError(t, err)
}

func TestCreateDummySavFile(t *testing.T) {
	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)
	defer helpers.CleanupTestSaveDirectory(tempDir)

	filePath := filepath.Join(tempDir, "test.sav")
	err = helpers.CreateDummySaveFile(filePath, helpers.TypeSav, helpers.SizeSmall)
	require.NoError(t, err)

	// Verify file exists and has expected size
	size, err := helpers.GetFileSize(filePath)
	require.NoError(t, err)
	assert.Equal(t, int64(helpers.SizeSmall), size)
}

func TestCreateRandomSaveFile(t *testing.T) {
	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)
	defer helpers.CleanupTestSaveDirectory(tempDir)

	for i := 0; i < 5; i++ {
		filePath, err := helpers.CreateRandomSaveFile(tempDir)
		require.NoError(t, err)

		// Verify file exists
		size, err := helpers.GetFileSize(filePath)
		require.NoError(t, err)
		assert.Greater(t, size, int64(0))

		// Verify file extension
		ext := filepath.Ext(filePath)
		assert.Contains(t, []string{".zip", ".sav"}, ext)

		// If it's a ZIP file, validate it
		if ext == ".zip" {
			err = helpers.ValidateZipFile(filePath)
			assert.NoError(t, err)
		}
	}
}

func TestFileSizeVariations(t *testing.T) {
	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)
	defer helpers.CleanupTestSaveDirectory(tempDir)

	sizes := []helpers.FileSize{
		helpers.SizeSmall,
		helpers.SizeMedium,
		helpers.SizeLarge,
	}

	for _, expectedSize := range sizes {
		t.Run(string(rune(expectedSize)), func(t *testing.T) {
			filePath := filepath.Join(tempDir, "test_size.sav")
			err := helpers.CreateDummySaveFile(filePath, helpers.TypeSav, expectedSize)
			require.NoError(t, err)

			actualSize, err := helpers.GetFileSize(filePath)
			require.NoError(t, err)
			assert.Equal(t, int64(expectedSize), actualSize)
		})
	}
}

func TestZipFileContents(t *testing.T) {
	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)
	defer helpers.CleanupTestSaveDirectory(tempDir)

	filePath := filepath.Join(tempDir, "test_contents.zip")
	err = helpers.CreateDummySaveFile(filePath, helpers.TypeZip, helpers.SizeLarge)
	require.NoError(t, err)

	// The ValidateZipFile function internally checks if files can be read
	err = helpers.ValidateZipFile(filePath)
	assert.NoError(t, err)
}

func TestInvalidFileType(t *testing.T) {
	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)
	defer helpers.CleanupTestSaveDirectory(tempDir)

	filePath := filepath.Join(tempDir, "test.invalid")
	err = helpers.CreateDummySaveFile(filePath, "invalid", helpers.SizeSmall)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
}

func TestFileCleanup(t *testing.T) {
	tempDir, err := helpers.CreateTestSaveDirectory()
	require.NoError(t, err)

	// Create some files
	filePath1 := filepath.Join(tempDir, "test1.zip")
	filePath2 := filepath.Join(tempDir, "test2.sav")

	err = helpers.CreateDummySaveFile(filePath1, helpers.TypeZip, helpers.SizeSmall)
	require.NoError(t, err)

	err = helpers.CreateDummySaveFile(filePath2, helpers.TypeSav, helpers.SizeSmall)
	require.NoError(t, err)

	// Verify files exist
	_, err = helpers.GetFileSize(filePath1)
	assert.NoError(t, err)

	_, err = helpers.GetFileSize(filePath2)
	assert.NoError(t, err)

	// Cleanup directory
	err = helpers.CleanupTestSaveDirectory(tempDir)
	assert.NoError(t, err)

	// Verify files are gone
	_, err = helpers.GetFileSize(filePath1)
	assert.Error(t, err)

	_, err = helpers.GetFileSize(filePath2)
	assert.Error(t, err)
}
