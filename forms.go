package aurora

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	valid "github.com/asaskevich/govalidator"
	"github.com/bluele/gforms"
)

var (
	MsgRequired  = "hili eneo halitakiwi kuachwa wazi"
	MsgName      = "hili eneo linatakiwa liwe mchanganyiko wa herufi na namba"
	MsgEmail     = "email sio sahihi. mfano gernest@aurora.com"
	MsgPhone     = "namba ya simu sio sahihi. mfano +25571700112233"
	MsgMinlength = "namba ya siri inatakiwa kuanzia herufi 6 na kuendelea"
	MsgEqual     = "%s inatakiwa iwe sawa na %s"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type validateFunc func(string) bool

type CustomValidator struct {
	Vf      validateFunc
	Message string
	gforms.Validator
}

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
				gforms.MinLengthValidator(6, MsgMinlength),
			},
		),
		gforms.NewTextField(
			"confirm_pass",
			gforms.Validators{
				gforms.Required(MsgRequired),
				IsName(),
				gforms.MinLengthValidator(6, MsgMinlength),
				EqualValidator{to: "pass", Message: MsgEqual},
			},
		),
	))
}

func IsName() CustomValidator {
	return CustomValidator{Vf: valid.IsAlphanumeric, Message: MsgName}
}

type EqualValidator struct {
	CustomValidator
	to      string
	Message string
}

func (vl EqualValidator) Validate(fi *gforms.FieldInstance, fo *gforms.FormInstance) error {
	v := fi.V
	if v.IsNil || v.Kind != reflect.String || v.Value == "" {
		return nil
	}
	fi2, ok := fo.GetField(vl.to)
	if !ok {
		log.Println("here")
		return fmt.Errorf("%s haipo", fi2.GetName())
	}
	v2 := fi2.GetV()

	if v.Value != v2.Value {
		return fmt.Errorf(vl.Message, fi.GetName(), fi2.GetName())
	}
	return nil

}

func ComposeLoginForm() gforms.Form {
	return gforms.DefineForm(gforms.NewFields(
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
				gforms.MinLengthValidator(6, MsgMinlength),
			},
		),
	))
}
