package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	discordWebhook string = "https://discordapp.com/api/webhooks/512929277923164180/1nu3FzCoPvsH3FevL3HkY9k6fooR6AIg7Bihn00-bEE2TILzmYfj-ZV22v-2sTQmdZy3"
)

/*
Notifies discord with a nice message
*/
func notifyDiscord(count int) {
	content := make(map[string]string) // Content for discord

	content["content"] = fmt.Sprintf("Amount of users in the database: %d", count)
	jsonResp, err := json.Marshal(content)
	if err != nil {
		fmt.Println("Error marshaling JSON")
	}

	_, err = http.Post(discordWebhook, "application/json", bytes.NewBuffer(jsonResp))
	if err != nil {
		fmt.Println("Error making POST request to discord.")
	}
}

/*
Checks every 10 seconds for a change
*/
func clockTrigger() {
	delay := time.Second * 10
	for {
		response := make(map[string]interface{}) // Map of the response from the API
		resp, err := http.Get("https://glacial-bastion-87425.herokuapp.com/api/usercount")
		if err != nil {
			fmt.Println("Error making GET request.")
			return
		}

		fmt.Println(resp.Body)
		json.NewDecoder(resp.Body).Decode(&response)
		fmt.Println(response)
		//count := response["count"].(int)

		notifyDiscord(1)
		time.Sleep(delay)
	}
}

func main() {
	clockTrigger()
}
