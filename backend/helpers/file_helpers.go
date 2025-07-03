package helpers

import (
	"archive/zip"
	"bytes"
)

// CreateDummyZip creates an in-memory zip file for testing purposes.
func CreateDummyZip() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Create a dummy file inside the zip archive
	f, err := w.Create("dummy.txt")
	if err != nil {
		return nil, err
	}
	_, err = f.Write([]byte("This is a dummy file."))
	if err != nil {
		return nil, err
	}

	// Close the writer to finalize the zip file
	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}