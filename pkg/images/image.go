package images

import (
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
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
