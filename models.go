package aurora

import "time"

const (
	male   = iota + 1 // 1
	female            // 2
	zombie            // 3
)

// Account is an interface for a user account
type Account interface {
	Email() string
	Password() string
}

// User contains details about a user
type User struct {
	UUID         string    `json:"uuid" gforms:"-"`
	FirstName    string    `json:"first_name" gforms:"first_name"`
	LastName     string    `json:"last_name" gforms:"last_name"`
	EmailAddress string    `json:"email" gforms:"email_address"`
	Pass         string    `json:"password" gforms:"pass"`
	ConfirmPass  string    `json:"-" gforms:"confirm_pass"`
	CreatedAt    time.Time `json:"created_at" gforms:"-"`
	UpdatedAt    time.Time `json:"updated_at" gforms:"-"`
}

// Email user email address
func (u *User) Email() string {
	return u.EmailAddress
}

// Password user password
func (u *User) Password() string {
	return u.Pass
}

// Profile contains additional information about the user
type Profile struct {
	ID        string    `json:"id" gforms:"-"`
	FirstName string    `json:"first_name" gforms:"first_name"`
	LastName  string    `json:"last_name" gforms:"last_name"`
	Picture   *Photo    `json:"picture" gforms:"-"`
	Age       int       `json:"age" gforms:"age"`
	IsUpdate  bool      `json:"is_update" gforms:"-"`
	BirthDate time.Time `json:"birth_date" gforms:"birth_date"`
	Gender    int       `json:"gender" gforms:"gender"`
	Photos    []*Photo  `json:"photos" gforms:"-"`
	City      string    `json:"city" gforms:"city"`
	Country   string    `json:"country" gforms:"country"`
	Street    string    `json:"street" gforms:"street"`
	CreatedAt time.Time `json:"created_at" gforms:"-"`
	UpdatedAt time.Time `json:"update_at" gforms:"-"`
}

func (p *Profile) MyBirthDay() string {
	t := time.Time{}
	if p.BirthDate.String() == t.String() {
		return time.Now().Format(birthDateFormat)
	}
	return p.BirthDate.Format(birthDateFormat)
}

func (p *Profile) Sex() string {
	switch p.Gender {
	case male:
		return "Mwanaume"
	case female:
		return "Mwanamke"
	case zombie:
		return "undead"
	}
	return ""
}
