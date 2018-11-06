package web

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_Register(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(HandlerRegister)) // TODO: So it doesnt add to the actual database
	defer testServer.Close()

	expected := "User created!"

	form := url.Values{}
	form.Add("username", "teser name")
	form.Add("email", "emailw@email.com")
	form.Add("password", "Password")
	form.Add("confirm", "Password")

	resp, err := http.PostForm(testServer.URL, form)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading the body: %s", err.Error())
	}

	respStr := string(respB)
	actualResponse := respStr[len(respStr)-14 : len(respStr)-1] // Now this is what I call a shitty way to do it (the response is an entire html document with that at the end)

	if actualResponse != expected {
		t.Errorf("Expected '%s' differs from actual '%s'.", expected, actualResponse)
	}
}
