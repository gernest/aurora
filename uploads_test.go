package aurora

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"testing"
)

func TestGetFileUpload(t *testing.T) {
	req, err := requestWithFile()
	if err != nil {
		t.Error(err)
	}
	f, err := GetFileUpload(req, "profile")
	if err != nil {
		t.Error(err)
	}
	checkExtension(f, "jpg", t)
	f, err = GetFileUpload(req, "nothere")
	if err == nil {
		t.Error("Expected an error, got nil instead")
	}
	if f != nil {
		t.Errorf("Expected nil, got %v", f)
	}
}

func TestGetMultipleFileUpload(t *testing.T) {
	req, err := requestMuliFile()
	if err != nil {
		t.Error(err)
	}
	files, err := GetMultipleFileUpload(req, "photos")
	if err != nil {
		list := err.(listErr)
		if len(list) != 2 {
			t.Errorf("Expected two errors got %d", len(list))
		}
		if len(files) != 3 {
			t.Errorf("Expected 3 files got %d", len(files))
		}
		if len(files) == 3 {
			xt := "jpg"
			for _, v := range files {
				checkExtension(v, xt, t)
			}
		}
	}
	if files != nil {
		if len(files) != 3 {
			t.Errorf("Expected 3 files got %d", len(files))
		}
		if len(files) == 3 {
			xt := "jpg"
			for _, v := range files {
				checkExtension(v, xt, t)
			}
		}
	}
	files, err = GetMultipleFileUpload(req, "nothere")
	if err == nil {
		t.Error("Expected an error, got nil instead")
	}

	req1, err := requestMultiWithoutErr()
	if err != nil {
		t.Error(err)
	}
	files, err = GetMultipleFileUpload(req1, "photos")
	if err != nil {
		t.Error(err)
	}
	if len(files) != 3 {
		t.Errorf("Expected 3 files got %d", len(files))
	}
}
func checkExtension(f *fileUpload, ext string, t *testing.T) {
	rext, err := getFileExt(*f.Body)
	if err != nil {
		t.Error(err)
	}
	if rext != ext {
		t.Errorf("Expected %s got %s", ext, rext)
	}
}

func requestWithFile() (*http.Request, error) {
	buf := new(bytes.Buffer)
	f, err := ioutil.ReadFile("public/img/me.jpg")
	if err != nil {
		return nil, err
	}
	w := multipart.NewWriter(buf)
	defer w.Close()
	ww, err := w.CreateFormFile("profile", "me.jpg")
	if err != nil {
		return nil, err
	}
	ww.Write(f)
	req, err := http.NewRequest("POST", "http://bogus.com", buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

func requestMuliFile() (*http.Request, error) {
	buf := new(bytes.Buffer)
	f, err := ioutil.ReadFile("public/img/me.jpg")
	if err != nil {
		return nil, err
	}
	w := multipart.NewWriter(buf)
	defer w.Close()
	first, err := w.CreateFormFile("photos", "home.jpg")
	if err != nil {
		return nil, err
	}
	first.Write(f)
	second, err := w.CreateFormFile("photos", "baby.jpg")
	if err != nil {
		return nil, err
	}
	second.Write(f)
	third, err := w.CreateFormFile("photos", "wanker.jpg")
	if err != nil {
		return nil, err
	}
	third.Write(f)
	fourth, err := w.CreateFormFile("photos", "wankerer.jpg")
	if err != nil {
		return nil, err
	}
	fourth.Write([]byte("shit"))

	fifth, err := w.CreateFormFile("photos", "wankeroma.jpg")
	if err != nil {
		return nil, err
	}
	fifth.Write([]byte("shit"))
	req, err := http.NewRequest("POST", "http://bogus.com", buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

func requestMultiWithoutErr() (*http.Request, error) {
	buf := new(bytes.Buffer)
	f, err := ioutil.ReadFile("public/img/me.jpg")
	if err != nil {
		return nil, err
	}
	w := multipart.NewWriter(buf)
	defer w.Close()
	first, err := w.CreateFormFile("photos", "home.jpg")
	if err != nil {
		return nil, err
	}
	first.Write(f)
	second, err := w.CreateFormFile("photos", "baby.jpg")
	if err != nil {
		return nil, err
	}
	second.Write(f)
	third, err := w.CreateFormFile("photos", "wanker.jpg")
	if err != nil {
		return nil, err
	}
	third.Write(f)
	req, err := http.NewRequest("POST", "http://bogus.com", buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}
