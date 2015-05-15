package aurora

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bluele/gforms"
)

func TestIsName(t *testing.T) {
	var (
		req1, req2 *http.Request
		err        error
		form       *gforms.FormInstance
		vars       url.Values
	)
	vars = url.Values{
		"name": {"gernest"},
	}
	Form := gforms.DefineForm(gforms.NewFields(
		gforms.NewTextField(
			"name",
			gforms.Validators{
				IsName(),
			},
		),
	))

	// Should pass when the name field is alpanumeric
	req1, err = http.NewRequest("POST", "/", strings.NewReader(vars.Encode()))
	if err != nil {
		t.Error(err)
	}
	req1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form = Form(req1)
	if !form.IsValid() {
		t.Error("Not expected: validation error.")
	}

	// Should fail when the name field aint alphanumeric
	vars = url.Values{
		"name": {"g-ernest"},
	}
	req2, err = http.NewRequest("POST", "/", strings.NewReader(vars.Encode()))
	if err != nil {
		t.Error(err)
	}
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form = Form(req2)
	if form.IsValid() {
		t.Error("Expected: validation error.")
	}
	if form.Errors().Get("name")[0] != MsgName {
		t.Errorf("Expected %s got %s", MsgName, form.Errors().Get("name")[0])
	}
}

func TestComposeRegisterForm(t *testing.T) {
	Form := ComposeRegisterForm()
	vars := url.Values{
		"first_name":    {"gernest"},
		"last_name":     {"aurora"},
		"email_address": {"gernest@aurora.com"},
		"pass":          {"mypassword"},
		"confirm_pass":  {"mypassword"},
	}
	req1, _ := http.NewRequest("POST", "/", strings.NewReader(vars.Encode()))
	req1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form1 := Form(req1)
	if !form1.IsValid() {
		t.Error("Expected form to be valid")
	}
	usr := form1.GetModel().(User)
	if usr.EmailAddress != vars.Get("email_address") {
		t.Errorf("Expected to get a user struct")
	}

	vars.Set("first_name", "---")
	req2, _ := http.NewRequest("POST", "/", strings.NewReader(vars.Encode()))
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form2 := Form(req2)
	if form2.IsValid() {
		t.Error("Expected: validation error.")
	}
}

func TestBirthDateValidator(t *testing.T) {
	var (
		req1, req2 *http.Request
		vars       url.Values
		err        error
		form       *gforms.FormInstance
		now        = time.Now()
		yearsAgo   = func(yrs int) time.Time {
			n := time.Now()
			nowAFter := n.AddDate(18, 1, 1)
			dur := nowAFter.Sub(n)
			return n.Add(-dur)

		}
	)
	Form := gforms.DefineForm(gforms.NewFields(
		gforms.NewDateTimeField(
			"date",
			time.RFC822,
			gforms.Validators{
				BirthDateValidator{Limit: 18, Message: MsgMinAge},
			},
		),
	))

	vars = url.Values{
		"date": {now.Format(time.RFC822)},
	}
	req1, err = http.NewRequest("POST", "/", strings.NewReader(vars.Encode()))
	if err != nil {
		t.Error(err)
	}
	req1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form = Form(req1)
	if form.IsValid() {
		t.Error("Expected some errors")
	}

	vars = url.Values{
		"date": {yearsAgo(18).Format(time.RFC822)},
	}
	req2, err = http.NewRequest("POST", "/", strings.NewReader(vars.Encode()))
	if err != nil {
		t.Error(err)
	}
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form = Form(req2)
	if !form.IsValid() {
		t.Error(form.Errors())
	}
}
