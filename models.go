package aurora

import "time"

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

// NewUser creates a new user and assings him a new uuid
func NewUser() *User {
	return &User{UUID: getUUID()}
}

// Profile contains additional information about the user
type Profile struct {
	ID        string    `json:"id" gforms:"-"`
	Picture   *Photo    `json:"picture" gforms:"-"`
	Age       int       `json:"age" gforms:"age"`
	BirthDate time.Time `json:"birth_date" gforms:"birth_date"`
	Height    int       `json:"height" gforms:"height"`
	Weight    int       `json:"weight" gforms:"weight"`
	Hobies    []string  `json:"hobies" gforms:"hobbies"`
	Photos    []*Photo  `json:"photos" gforms:"-"`
	City      string    `json:"city" gforms:"city"`
	Country   string    `json:"country" gforms:"country"`
	Street    string    `json:"street" gforms:"street"`
	CreatedAt time.Time `json:"created_at" gforms:"-"`
	UpdatedAt time.Time `json:"update_at" gforms:"-"`
}
