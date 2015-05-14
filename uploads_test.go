package aurora

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"testing"
)

func TestGetFileUpload(t *testing.T) {
	req, err := requestWithFile("me.jpg")
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

	req2, err := requestWithFile("mint.png")
	if err != nil {
		t.Error(err)
	}
	f, err = GetFileUpload(req2, "profile")
	if err != nil {
		t.Error(err)
	}
	checkExtension(f, "png", t)

}

func TestGetMultipleFileUpload(t *testing.T) {
	req := requestMuliFile("me.jpg", t)
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

	// just a bonus, Wanna know if listErr is fine
	testListErr(t)
}
func TestSaveUploadFile(t *testing.T) {
	var (
		pBucket = "profiles"
	)
	// JPG
	id := "db0668ac-7eba-40dd-56ee-0b1c0b9b415p"
	pdb := setDB(testDb, "fixture/uploads.bdb")
	defer pdb.DeleteDatabase()
	req, err := requestWithFile("me.jpg")
	if err != nil {
		t.Error(err)
	}
	f, err := GetFileUpload(req, "profile")
	if err != nil {
		t.Error(err)
	}
	checkExtension(f, "jpg", t)

	err = CreateProfile(pdb, &Profile{ID: id}, pBucket)
	if err != nil {
		t.Error(err)
	}
	p, err := GetProfile(pdb, pBucket, id)
	if err != nil {
		t.Error(err)
	}
	pic, err := SaveUploadFile(pdb, f, p)
	if err != nil {
		t.Error(err)
	}
	if f.Ext != pic.Type {
		t.Errorf("Expected %s  got %s", f.Ext, pic.Type)
	}

	// PNG
	req2, err := requestWithFile("mint.png")
	if err != nil {
		t.Error(err)
	}
	f, err = GetFileUpload(req2, "profile")
	if err != nil {
		t.Error(err)
	}
	checkExtension(f, "png", t)
	pic, err = SaveUploadFile(pdb, f, p)
	if err != nil {
		t.Error(err)
	}
	if f.Ext != pic.Type {
		t.Errorf("Expected %s  got %s", f.Ext, pic.Type)
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

func requestWithFile(fileName string) (*http.Request, error) {
	buf := new(bytes.Buffer)
	public := "public/img/"
	f, err := ioutil.ReadFile(fmt.Sprintf("%s%s", public, fileName))
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

func requestMuliFile(fileName string, t *testing.T) *http.Request {
	buf, cType := testUpData(fileName, "multi", t)
	req, err := http.NewRequest("POST", "http://bogus.com", buf)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", cType)
	return req
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

func testListErr(t *testing.T) {
	var err listErr
	hello := errors.New("hello")
	world := errors.New("wordl")
	err = append(err, hello, world)
	if err.Error() != hello.Error()+", "+world.Error() {
		t.Errorf("Expected %s, %s got %s", hello.Error(), world.Error(), err.Error())
	}
}

func testUpData(fileName, kind string, t *testing.T) (*bytes.Buffer, string) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	public := "public/img/"
	defer w.Close()
	switch kind {
	case "multi":
		f, err := ioutil.ReadFile(fmt.Sprintf("%s%s", public, fileName))
		if err != nil {
			t.Error(err)
		}
		first, err := w.CreateFormFile("photos", "home.jpg")
		if err != nil {
			t.Error(err)
		}
		first.Write(f)
		second, err := w.CreateFormFile("photos", "baby.jpg")
		if err != nil {
			t.Error(err)
		}
		second.Write(f)
		third, err := w.CreateFormFile("photos", "wanker.jpg")
		if err != nil {
			t.Error(err)
		}
		third.Write(f)
		fourth, err := w.CreateFormFile("photos", "wankerer.jpg")
		if err != nil {
			t.Error(err)
		}
		fourth.Write([]byte("shit"))

		fifth, err := w.CreateFormFile("photos", "wankeroma.jpg")
		if err != nil {
			t.Error(err)
		}
		fifth.Write([]byte("shit"))
	case "single":
		f, err := ioutil.ReadFile(fmt.Sprintf("%s%s", public, fileName))
		if err != nil {
			t.Error(err)
		}
		ww, err := w.CreateFormFile("profile", "me.jpg")
		if err != nil {
			t.Error(err)
		}
		ww.Write(f)

	}
	return buf, w.FormDataContentType()
}
