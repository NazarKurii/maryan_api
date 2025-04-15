package images

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func Save(path string, image *multipart.FileHeader) error {
	src, err := image.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file failed: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, src)
	if err != nil {
		return fmt.Errorf("copying file data failed: %w", err)
	}
	fmt.Printf("Image saved at %s (%d bytes)\n", path, written)
	return err
}
