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

func Test_Login(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerRegister))
	defer testServer.Close()

	username := time.Now().String()

	form := url.Values{}
	form.Add("username", username)
	form.Add("email", time.Now().String())
	form.Add("password", "Password")
	form.Add("confirm", "Password")

	resp, err := http.PostForm(testServer.URL, form) // Add a user first
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Statuscode expected to be %d, but is %d.", http.StatusCreated, resp.StatusCode)
	}

	testServer = httptest.NewServer(http.HandlerFunc(HandlerLogin)) // Create a new server with the correct handler
	form = url.Values{}
	form.Add("username", username)
	form.Add("password", "Password")

	resp, err = http.PostForm(testServer.URL, form) // Try to log in with the newly created user
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code expected to be %d, but is %d.", http.StatusOK, resp.StatusCode)
	}
}
