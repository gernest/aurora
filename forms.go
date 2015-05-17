package aurora

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/bluele/gforms"
)

const ageLimit int = 18

var (
	// MsgRequired is the error message for required validation.
	MsgRequired = "hili eneo halitakiwi kuachwa wazi"

	// MsgName is the error message displayed for a name validation
	MsgName = "hili eneo linatakiwa liwe mchanganyiko wa herufi na namba"

	//MsgEmail is the error message displayed for email validation
	MsgEmail = "email sio sahihi. mfano gernest@aurora.com"

	// MsgMinLength is the error message for a minimum length validation
	MsgMinLength = "namba ya siri inatakiwa kuanzia herufi 6 na kuendelea"

	// MsgEqual is the error message for equality validation
	MsgEqual = "%s inatakiwa iwe sawa na %s"

	// MsgMinAge the minimum age linit
	MsgMinAge = "umri unatakiwa uwe zaidi ya miaka %d"
)

// This is an interface which is helpful for implementing a custom validator.
// I adopted this so that I can use the functions defined in the govalidator library.
type validateFunc func(string) bool

// CustomValidator a custom validator for gforms
type CustomValidator struct {
	Vf      validateFunc
	Message string
	gforms.Validator
}

// Validate validates fields
func (vl CustomValidator) Validate(fi *gforms.FieldInstance, fo *gforms.FormInstance) error {
	v := fi.V
	if v.IsNil || v.Kind != reflect.String || v.Value == "" {
		return nil
	}
	if !vl.Vf(v.RawStr) {
		return errors.New(vl.Message)
	}
	return nil

}

// ComposeRegisterForm builds a registration form for validation(with gforms)
func ComposeRegisterForm() gforms.ModelForm {
	return gforms.DefineModelForm(User{}, gforms.NewFields(
		gforms.NewTextField(
			"first_name",
			gforms.Validators{
				gforms.Required(MsgRequired),
				IsName(),
			},
		),
		gforms.NewTextField(
			"last_name",
			gforms.Validators{
				gforms.Required(MsgRequired),
				IsName(),
			},
		),
		gforms.NewTextField(
			"email_address",
			gforms.Validators{
				gforms.Required(MsgRequired),
				gforms.EmailValidator(MsgEmail),
			},
		),
		gforms.NewTextField(
			"pass",
			gforms.Validators{
				gforms.Required(MsgRequired),
				IsName(),
				gforms.MinLengthValidator(6, MsgMinLength),
			},
		),
		gforms.NewTextField(
			"confirm_pass",
			gforms.Validators{
				gforms.Required(MsgRequired),
				IsName(),
				gforms.MinLengthValidator(6, MsgMinLength),
				EqualValidator{to: "pass", Message: MsgEqual},
			},
		),
	))
}

// IsName retruns a name validator
func IsName() CustomValidator {
	return CustomValidator{Vf: valid.IsAlphanumeric, Message: MsgName}
}

// EqualValidator checks if the two fields are equal. The to attribute is the name of
// the field whose value must be equal to the current field
type EqualValidator struct {
	gforms.Validator
	to      string
	Message string
}

// Validate  checks if the given field is egual to the field in the to attribute
func (vl EqualValidator) Validate(fi *gforms.FieldInstance, fo *gforms.FormInstance) error {
	v := fi.V
	if v.IsNil || v.Kind != reflect.String || v.Value == "" {
		return nil
	}
	fi2, ok := fo.GetField(vl.to)
	if ok {
		v2 := fi2.GetV()

		if v.Value != v2.Value {
			return fmt.Errorf(vl.Message, fi.GetName(), fi2.GetName())
		}
	}
	return nil
}

// BirthDateValidator validates the birth date, handy to keep minors offsite
type BirthDateValidator struct {
	Limit   int
	Message string
	gforms.Validator
}

// Validate checks if the given field instance esceeds the Limit attribute
func (vl BirthDateValidator) Validate(fi *gforms.FieldInstance, fo *gforms.FormInstance) error {
	v := fi.V
	if v.IsNil {
		return nil
	}
	iv := v.Value.(time.Time)
	now := time.Now()
	if now.Year()-iv.Year() < vl.Limit {
		return fmt.Errorf(vl.Message, vl.Limit)
	}
	return nil
}

// holds login form data
type loginForm struct {
	Email    string `gforms:"email"`
	Password string `gforms:"password"`
}

// ComposeLoginForm builds a login form for validation( with gforms)
func ComposeLoginForm() gforms.ModelForm {
	return gforms.DefineModelForm(loginForm{}, gforms.NewFields(
		gforms.NewTextField(
			"email",
			gforms.Validators{
				gforms.Required(MsgRequired),
				gforms.EmailValidator(MsgEmail),
			},
		),
		gforms.NewTextField(
			"password",
			gforms.Validators{
				gforms.Required(MsgRequired),
				gforms.MinLengthValidator(6, MsgMinLength),
			},
		),
	))
}

// ComposeProfileForm builds a profile form for validation (using gform)
func ComposeProfileForm() gforms.ModelForm {
	return gforms.DefineModelForm(Profile{}, gforms.NewFields(
		gforms.NewIntegerField(
			"age",
			gforms.Validators{
				gforms.MinValueValidator(ageLimit, MsgMinAge),
			},
		),
		gforms.NewDateTimeField(
			"birth_date",
			time.RFC822,
			gforms.Validators{
				gforms.Required(MsgRequired),
				BirthDateValidator{Limit: ageLimit, Message: MsgMinAge},
			},
		),
	))
}
