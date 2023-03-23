package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const city = "Oslo"

type OpenWeatherMapResponse struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Sys struct {
		Sunrise int64 `json:"sunrise"`
		Sunset  int64 `json:"sunset"`
	} `json:"sys"`
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
}

type SlackMessage struct {
	Text string `json:"text"`
}

func spring(data *OpenWeatherMapResponse) bool {
	currentTime := time.Now().Unix()
	daylightHours := data.Sys.Sunset - data.Sys.Sunrise
	isClear := data.Weather[0].Main == "Clear"
	isWarm := data.Main.Temp > 10
	isDaytime := currentTime > data.Sys.Sunrise && currentTime < data.Sys.Sunset
	return isClear && isWarm && daylightHours > 43200 && isDaytime
}

func sendSlackNotification(message SlackMessage, webhookURL string) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = http.PostForm(webhookURL, url.Values{"payload": {string(payload)}})
	return err
}

func main() {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		log.Fatal(errors.New("OPENWEATHERMAP_API_KEY environment variable not set"))
	}

	response, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city + ",no&appid=" + apiKey + "&units=metric")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data OpenWeatherMapResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
	}

	if spring(&data) {
		webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
		if webhookURL == "" {
			log.Fatal(errors.New("SLACK_WEBHOOK_URL environment variable not set"))
		}
		message := SlackMessage{Text: "It's springtime in " + city + "!"}
		err = sendSlackNotification(message, webhookURL)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Notification sent to Slack!")
	} else {
		fmt.Println("It's not spring yet.")
	}
}
