package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func Test_Register(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerRegister)) // TODO: So it doesnt add to the actual database
	defer testServer.Close()

	expected := http.StatusCreated

	form := url.Values{}
	form.Add("username", time.Now().String()) // To not add a user with the same credentials for each test
	form.Add("email", time.Now().String())
	form.Add("password", "Password")
	form.Add("confirm", "Password")

	resp, err := http.PostForm(testServer.URL, form)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if resp.StatusCode != expected {
		t.Errorf("Statuscode expected to be %d, but is %d.", expected, resp.StatusCode)
	}
}
