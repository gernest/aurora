package aurora

import (
	"os"
	"testing"
)

func TestFlash(t *testing.T) {
	var (
		flash   = NewFlash()
		success = "success"
		notice  = "note"
		err     = "error"
	)

	defer clenUp(t)
	flash.Success(success)
	flash.Notice(notice)
	flash.Error(err)
	d := flash.Data
	if d["FlashNotice"].(string) != notice {
		t.Errorf("Expected %s got %s", notice, d["FlashNotice"])
	}
	if d["FlashSuccess"].(string) != success {
		t.Errorf("Expected %s got %s", notice, d["FlashSuccess"])
	}
	if d["FlashError"].(string) != err {
		t.Errorf("Expected %s got %s", err, d["FlashError"])
	}
}

// deletes test database files
func clenUp(t *testing.T) {
	ts, _, rx := testServer(t)
	defer ts.Close()
	err := os.RemoveAll(rx.cfg.DBDir)
	if err != nil {
		t.Error(err)
	}
}
