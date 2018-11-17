package webhook

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// MessageDiscord notifys discord of new user
func MessageDiscord() {
	url := "https://discordapp.com/api/webhooks/512929277923164180/1nu3FzCoPvsH3FevL3HkY9k6fooR6AIg7Bihn00-bEE2TILzmYfj-ZV22v-2sTQmdZy3"

	// send message to discord
	s := `{"content":"New user have been added."}`

	var jsonStr = []byte(s)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}
