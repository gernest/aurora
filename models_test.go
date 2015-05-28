package aurora

import (
	"strings"
	"testing"
	"time"
)

func TestProfile_MyBirthDay(t *testing.T) {
	n := time.Now()
	p := &Profile{
		BirthDate: n,
	}
	if p.MyBirthDay() != n.Format(birthDateFormat) {
		t.Errorf("expected %s got %s", n.Format(birthDateFormat), p.MyBirthDay())
	}
}

func TestProfile_Sex(t *testing.T) {
	p := &Profile{}

	if p.Sex() != "" {
		t.Errorf("expected a empty string got %s", p.Sex())
	}
	p.Gender = male
	if strings.ToLower(p.Sex()) != "mwanaume" {
		t.Errorf("expected  mwanaume got %s", strings.ToLower(p.Sex()))
	}
	p.Gender = female
	if strings.ToLower(p.Sex()) != "mwanamke" {
		t.Errorf("expected mwanamke got %s", strings.ToLower(p.Sex()))
	}
	p.Gender = zombie
	if strings.ToLower(p.Sex()) != "undead" {
		t.Errorf("expected undead got %s", strings.ToLower(p.Sex()))
	}
}
