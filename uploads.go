package aurora

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
)

const defaultMaxMemory = 32 << 20 //32MB

type fileUpload struct {
	Body *multipart.File
	Ext  string
}

type photo struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Size       int       `json:"size"`
	UploadedBy string    `json:"uploaded_by"`
	UploadedAt time.Time `json:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func GetFileUpload(r *http.Request, fieldName string) (*fileUpload, error) {
	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return nil, err
	}
	return getUploadFile(file)
}

type listErr []error

func (l listErr) Error() string {
	var rst string
	for _, e := range l {
		rst = rst + ", " + e.Error()
	}
	return rst
}
func GetMultipleFileUpload(r *http.Request, fieldName string) ([]*fileUpload, error) {
	err := r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return nil, err
	}
	if up := r.MultipartForm.File[fieldName]; len(up) > 0 {
		var rst []*fileUpload
		var ferr listErr
		for _, v := range up {
			f, err := v.Open()
			if err != nil {
				ferr = append(ferr, err)
				continue
			}
			file, err := getUploadFile(f)
			if err != nil {
				ferr = append(ferr, err)
				continue
			}
			rst = append(rst, file)
		}
		if len(ferr) > 0 {
			return rst, ferr
		}
		return rst, nil
	}
	return nil, http.ErrMissingFile
}

func getFileExt(file multipart.File) (string, error) {
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	defer file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	f := http.DetectContentType(buf)
	switch f {
	case "image/jpeg", "image/jpg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	default:
		return "", fmt.Errorf("file %s not supported", f)
	}
}

func getUploadFile(file multipart.File) (*fileUpload, error) {
	ext, err := getFileExt(file)
	if err != nil {
		return nil, err
	}
	return &fileUpload{&file, ext}, nil
}
