package aurora

import (
	"testing"
)

func TestFlash(t *testing.T) {
	success := "success"
	notice := "note"
	err := "error"
	flash := NewFlash()
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
