package aurora

import "testing"

var (
	pids = []string{
		"db0668ac-7eba-40dd-56ee-0b1c0b9b415d",
		"e6917dfe-b4f6-49b8-5628-83dd2a430e9a",
		"bc5288cf-4120-4f3c-5957-b19e093a12f4",
	}
	pBucket = "profiles"
)

func TestCreateProfile(t *testing.T) {
	for _, id := range pids {
		p := &Profile{ID: id}
		err := CreateProfile(testDb, p, pBucket)
		if err != nil {
			t.Error(err)
		}
	}
	err := CreateProfile(testDb, &Profile{ID: pids[0]}, pBucket)
	if err == nil {
		t.Error("Expected an error")
	}
}
func TestGetProfile(t *testing.T) {
	for _, id := range pids {
		p, err := GetProfile(testDb, pBucket, id)
		if err != nil {
			t.Error(err)
		}
		if err == nil {
			if p.ID != id {
				t.Errorf("Expected %s got %s", id, p.ID)
			}
		}
	}
	p, err := GetProfile(testDb, pBucket, "bogus")
	if err == nil {
		t.Error("Expected an error")
	}
	if p != nil {
		t.Errorf("Expected nil, got %v", p)
	}
}
func TestUpdateProfile(t *testing.T) {
	var (
		city    = "mwanza"
		country = "Tanzania"
	)
	for _, id := range pids {
		p, err := GetProfile(testDb, pBucket, id)
		if err != nil {
			t.Error(err)
		}
		if err == nil {
			p.City = city
			p.Country = country
			err = UpdateProfile(testDb, p, pBucket)
			if err != nil {
				t.Error(err)
			}
		}
	}
	p := &Profile{ID: "bogus", Country: country, City: city}
	err := UpdateProfile(testDb, p, pBucket)
	if err == nil {
		t.Error("Expected an error got nil instead")
	}

}
