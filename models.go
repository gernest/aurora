package aurora

import (
	"time"

	"github.com/nu7hatch/gouuid"
)

type Account interface {
	Email() string
	Password() string
	VerifyPassword(string) bool
}

type User struct {
	UUID        string    `json:"uuid" gforsm:"-"`
	FirstName   string    `json:"first_name" gforms:"first_name"`
	LastName    string    `json:"last_name" gforms:"last_name"`
	EmailAdress string    `json:"email" gforms:"email_address"`
	Pass        string    `json:"password" gform:"pass"`
	ConfirmPass string    `json:"-" gforms:"confirm_pass"`
	CreatedAt   time.Time `json:"created_at" gforms:"-"`
	UpdatedAt   time.Time `json:"updated_at" gforms:"-"`
}

func (u *User) Email() string {
	return u.EmailAdress
}
func (u *User) Password() string {
	return u.Pass
}
func (u *User) VerifyPassword(pass string) bool {
	return true
}

func NewUser() *User {
	id, err := uuid.NewV4()
	if err != nil {
		// TODO: Log and try a new one
	}
	return &User{UUID: id.String()}
}

type Profile struct {
	ID        string    `json:"id"`
	Picture   string    `json:"picture"`
	Age       int       `json:"age"`
	BirthDate time.Time `json:"birth_date"`
	Height    int       `json:"height"`
	Weight    int       `json:"weight"`
	Hobies    []string  `json:"hobies"`
	Photos    []string  `json:"photos"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	Street    string    `json:"street"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"update_at"`
}
