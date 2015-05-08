package aurora

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/bluele/gforms"
)

func TestIsName(t *testing.T) {
	Form := gforms.DefineForm(gforms.NewFields(
		gforms.NewTextField(
			"name",
			gforms.Validators{
				IsName(),
			},
		),
	))
	req1, _ := http.NewRequest("POST", "/", strings.NewReader(url.Values{"name": {"gernest"}}.Encode()))
	req1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form1 := Form(req1)
	if !form1.IsValid() {
		t.Error("Not expected: validation error.")
	}
	req2, _ := http.NewRequest("POST", "/", strings.NewReader(url.Values{"name": {"gernest--"}}.Encode()))
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form2 := Form(req2)
	if form2.IsValid() {
		t.Error("Expected: validation error.")
	}
	if form2.Errors().Get("name")[0] != MsgName {
		t.Errorf("Expected %s got %s", MsgName, form2.Errors().Get("name")[0])
	}
}

func TestComposeRegisterForm(t *testing.T) {
	Form := ComposeRegisterForm()
	vars := url.Values{}
	vars.Set("first_name", "gernest")
	vars.Set("last_name", "aurora")
	vars.Set("email_address", "gernest@aurora.com")
	vars.Set("pass", "bogusBogus")
	vars.Set("confirm_pass", "bogusBogus")
	req1, _ := http.NewRequest("POST", "/", strings.NewReader(vars.Encode()))
	req1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form1 := Form(req1)
	if !form1.IsValid() {
		for k, v := range form1.Errors() {
			t.Errorf("%s ---> %s", k, v)
		}
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
