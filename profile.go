package aurora

import "github.com/gernest/nutz"

// CreateProfile creates a new profile using Profile.ID as the jey
func CreateProfile(db nutz.Storage, p *Profile, bucket string, nest ...string) error {
	return createIfNotexist(db, p, bucket, p.ID)
}

// GetProfile retrives a profile with a given id
func GetProfile(db nutz.Storage, bucket, id string, nest ...string) (*Profile, error) {
	p := &Profile{}
	err := getAndUnmarshall(db, bucket, id, p)
	if err != nil {
		return nil, err
	}
	return p, err
}

// UpdateProfile updates a given profile
func UpdateProfile(db nutz.Storage, p *Profile, bucket string, nest ...string) error {
	return marshalAndUpdate(db, p, bucket, p.ID)
}
