package aurora

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gernest/nutz"
)

// FileUpload represents the uploaded file
type FileUpload struct {
	Body *multipart.File
	Ext  string
}

// Photo is the metadata of an uploaded image file
type Photo struct {
	ID string `json:"id"`

	// Type is the photo's file extension e.g jpeg or png
	Type string `json:"type"`

	//Size is the size of the photo.
	Size int `json:"size"`

	//UploadedBy is the ID of the user who uploaded the photo
	UploadedBy string `json:"uploaded_by"`

	UploadedAt time.Time `json:"uploaded_at"`

	// UpdatedAt is the time the photo was updated. I keep this filed so as
	// to provide, last modified time when serving the photo.
	UpdatedAt time.Time `json:"updated_at"`
}

// GetFileUpload retrieves uploaded file from a request.This function, returns only
// the first file that matches, thus retrieving a single file only.
// the fieldName parameter is the name of the field which holds the file data.
func GetFileUpload(r *http.Request, fieldName string) (*FileUpload, error) {
	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return nil, err
	}
	return getUploadFile(file)
}

// a slice to hold a couple of errors
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

// SaveUploadFile saves the uploaded photos to the profile database. In aurora, every user
// has his/her own personal database.
//
// The db argument should be the user's database. The uploaded file is storesd in two versions
// meta, and data. The meta, is the metadata about the uploaded file, in our case a Photo
// object. The photo object is marshalled and stored in a metaBucket.
//
// The data part is the actual encoded file, its stored in the dataBucket.
func SaveUploadFile(db nutz.Storage, file *FileUpload, p *Profile) (*Photo, error) {
	var (
		// The bucket in which all photos will reside.
		photoBucket = "photos"

		// The bucket which stores metadata about the photos. This bucket iscreated
		// inside the photoBucket.
		metaBucket = "meta"

		// The bucket in which actual data that is in []byte is stored. its also created
		// inside the photoBucket
		dataBucket = "data"

		// NOTE: To keep the structure of recording data sane, I have used nested buckets.
		// So, the structure of the photo storage buckets is roughly like this.
		//
		// photoBucket
		//			 |---metaBucket
		//			 |---daaBucket
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

// extracts file extension.
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

// Returns  *FileUpload from the given multipart data. There is nothing fancy here, only that
// We need to get the file extension.
func getUploadFile(file multipart.File) (*FileUpload, error) {
	ext, err := getFileExt(file)
	if err != nil {
		return nil, err
	}
	return &FileUpload{&file, ext}, nil
}

// encodes a given photo, and returns a []byte of the photo. It currently supports
// png, and jpeg formats. The encoded data is the one which will be stored in the database.
func encodePhoto(file *FileUpload) ([]byte, error) {
	ext := file.Ext
	switch ext {
	case "jpg", "jpeg":
		img, err := jpeg.Decode(*file.Body)
		if err != nil {
			return nil, err
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
		buf := new(bytes.Buffer)
		png.Encode(buf, img)
		return buf.Bytes(), nil
	}
	return nil, errors.New("aurora: file not supported")
}
