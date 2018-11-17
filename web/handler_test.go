package web

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/jigurd/VirtualTabletop/tabletop"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

/*
Create a random string (used for email generation)
*/
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Test_Register(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerRegister))
	defer testServer.Close()

	expected := http.StatusCreated

	username := time.Now().String() // To not add a user with the same credentials for each test

	form := url.Values{}
	form.Add("username", username)
	form.Add("email", randSeq(5)+"@email.com")
	form.Add("password", "Password")
	form.Add("confirm", "Password")

	resp, err := http.PostForm(testServer.URL, form)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if resp.StatusCode != expected {
		t.Errorf("Statuscode expected to be %d, but is %d.", expected, resp.StatusCode)
	}

	if !tabletop.UserDB.Remove(username) { // The user is added to actual database, so remove it again
		fmt.Println("Warning: There was an error with deleting the user.")
	}
}

func Test_Login(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerRegister))
	defer testServer.Close()

	username := time.Now().String()

	form := url.Values{}
	form.Add("username", username)
	form.Add("email", randSeq(5)+"@email.com")
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

	if !tabletop.UserDB.Remove(username) {
		fmt.Println("Warning: There was an error with deleting the user.")
	}
}

/*
Tests that registering fails when the form sent is badly formed
*/
func Test_RegisterMalformedForm(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerRegister)) // TODO: So it doesnt add to the actual database
	defer testServer.Close()

	expected := http.StatusUnprocessableEntity

	form := url.Values{}
	form.Add("username1", time.Now().String()) // Badly formed form key
	form.Add("email", randSeq(5)+"@email.com")
	form.Add("password", "Password")
	form.Add("confirm", "Password")

	resp, err := http.PostForm(testServer.URL, form)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if resp.StatusCode != expected {
		t.Errorf("Statuscode expected to be %d, but is %d.", expected, resp.StatusCode)
	}

	form = url.Values{}
	form.Add("username", time.Now().String())
	form.Add("email", randSeq(5)+"@email.com")
	form.Add("password", "") // Badly formed form value
	form.Add("confirm", "Password")

	resp, err = http.PostForm(testServer.URL, form)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if resp.StatusCode != expected {
		t.Errorf("Statuscode expected to be %d, but is %d.", expected, resp.StatusCode)
	}
}
