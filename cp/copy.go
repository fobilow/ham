package cp

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Dir copies a whole directory recursively
func Dir(src string, dest string) error {
	// Get properties of the source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	// Create the destination dir
	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return err
	}

	// Read all contents of the directory
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	// Loop through all entries
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := Dir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err := File(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// File copies a single file from src to dest
func File(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	// Copy the file mode
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.Chmod(dest, srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}

func main() {
	srcDir := "path/to/source/directory"
	destDir := "path/to/destination/directory"

	if err := Dir(srcDir, destDir); err != nil {
		fmt.Printf("Error copying directory: %v\n", err)
	} else {
		fmt.Println("Directory copied successfully.")
	}
}
