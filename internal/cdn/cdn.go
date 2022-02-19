package cdn

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// ReceiveFileUpload handles the reception of uploads and puts the files into a specified location
func ReceiveFileUpload(ctx context.Context, req *http.Request, location, formName string) (string, error) {
	var path string

	mr, err := req.MultipartReader()
	if err != nil {
		return "", err
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if part.FormName() == formName {
			path = filepath.Join(location, part.FileName())

			os.MkdirAll(filepath.Dir(path), os.ModePerm) // make sure sub-folders exist
			out, err := os.Create(path)
			if err != nil {
				return "", err
			}
			defer out.Close()

			if _, err := io.Copy(out, part); err != nil {
				return "", err
			}
		}
	}

	return path, nil
}
