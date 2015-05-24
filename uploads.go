package aurora

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/gernest/nutz"
	"github.com/koyachi/go-nude"
)

var (
	errIsNude = errors.New("samahani picha za utupu haziruhusiwi")
)

// FileUpload represents the uploaded file
type FileUpload struct {
	Body *multipart.File
	Ext  string
}

// Photo is therepresentation of an uploaded image file
type Photo struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Size       int       `json:"size"`
	Query      string    `json:"query"`
	UploadedBy string    `json:"uploaded_by"`
	UploadedAt time.Time `json:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (p *Photo) GetQuery() template.HTML {
	return template.HTML(fmt.Sprintf("iid=%s&pid=%s", p.ID, p.UploadedBy))
}

// GetFileUpload retrieves uploaded file from a request.
// the fieldName parameter is the name of the field which holds the file data.
func GetFileUpload(r *http.Request, fieldName string) (*FileUpload, error) {
	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return nil, err
	}
	return getUploadFile(file)
}

type listErr []error

func (l listErr) Error() string {
	var rst string
	for i, e := range l {
		if i == 0 {
			if e != nil {
				rst = e.Error()
			}
			continue
		}
		if e != nil {
			rst = rst + ", " + e.Error()
		}
	}
	return rst
}

// GetMultipleFileUpload retrieves multiple files uploaded on a single request.
// The fieldName parameter is the form field containing the files
func GetMultipleFileUpload(r *http.Request, fieldName string) ([]*FileUpload, error) {
	const defaultMaxMemory = 32 << 20 //32MB

	err := r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return nil, err
	}
	if up := r.MultipartForm.File[fieldName]; len(up) > 0 {
		var rst []*FileUpload
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

// SaveUploadFile persists the given file
func SaveUploadFile(db nutz.Storage, file *FileUpload, p *Profile) (*Photo, error) {
	var (
		qPicID      = "iid"
		qProfID     = "pid"
		photoBucket = "photos"
		metaBucket  = "meta"
		dataBucket  = "data"
	)
	pic := &Photo{
		ID:         getUUID(),
		Type:       file.Ext,
		UploadedBy: p.ID,
		UploadedAt: time.Now(),
		UpdatedAt:  time.Now(),
	}
	data, err := encodePhoto(file)
	if err != nil {
		return nil, err
	}
	pic.Size = len(data)
	query := url.Values{
		qPicID:  {pic.ID},
		qProfID: {p.ID},
	}
	pic.Query = query.Encode()
	err = marshalAndCreate(db, pic, photoBucket, pic.ID, metaBucket)
	if err != nil {
		return nil, err
	}
	s := db.Create(photoBucket, pic.ID, data, dataBucket)
	if s.Error != nil {
		return nil, s.Error
	}
	return pic, nil
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
		return "", fmt.Errorf("aurora: file %s not supported", f)
	}
}

func getUploadFile(file multipart.File) (*FileUpload, error) {
	ext, err := getFileExt(file)
	if err != nil {
		return nil, err
	}
	return &FileUpload{&file, ext}, nil
}

func encodePhoto(file *FileUpload) ([]byte, error) {
	ext := file.Ext
	switch ext {
	case "jpg", "jpeg":
		img, err := jpeg.Decode(*file.Body)
		if err != nil {
			return nil, err
		}
		ok, err := nude.IsImageNude(img)
		if err != nil {
			// log this error
		}
		if ok {
			return nil, errIsNude
		}

		// this is supposed to increase the quality of the image. But I'm not sure
		// yet if it is necessary or we should just put nil, which will result into
		// using default values.
		opts := jpeg.Options{Quality: 98}

		buf := new(bytes.Buffer)
		jpeg.Encode(buf, img, &opts)
		return buf.Bytes(), nil
	case "png", "PNG":
		img, err := png.Decode(*file.Body)
		if err != nil {
			return nil, err
		}
		ok, err := nude.IsImageNude(img)
		if err != nil {
			// log this error
		}
		if ok {
			return nil, errIsNude
		}

		buf := new(bytes.Buffer)
		png.Encode(buf, img)
		return buf.Bytes(), nil
	}
	return nil, errors.New("aurora: file not supported")
}
