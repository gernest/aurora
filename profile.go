package aurora

import "github.com/gernest/nutz"

func CreateProfile(db nutz.Storage, p *Profile, bucket string, nest ...string) error {
	return createIfNotexist(db, p, bucket, p.ID)
}

func GetProfile(db nutz.Storage, bucket, id string, nest ...string) (*Profile, error) {
	p := &Profile{}
	err := getAndUnmarshall(db, bucket, id, p)
	if err != nil {
		return nil, err
	}
	return p, err
}

func UpdateProfile(db nutz.Storage, p *Profile, bucket string, nest ...string) error {
	return marshalAndUpdate(db, p, bucket, p.ID)
}
