package Commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SunsetResponse struct {
	Results struct {
		Sunset string `json:"sunset"`
	} `json:"results"`
	Status string `json:"status"`
}

type SunriseResponse struct {
	Results struct {
		Sunrise string `json:"sunrise"`
	} `json:"results"`
	Status string `json:"status"`
}

func GetSunset(latitude, longitude float64) (string, error) {

	url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%f&lng=%f&formatted=0", latitude, longitude)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var data SunsetResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	if data.Status != "OK" {
		return "", fmt.Errorf("API response status: %s", data.Status)
	}

	sunsetTime, err := time.Parse(time.RFC3339, data.Results.Sunset)
	if err != nil {
		return "", fmt.Errorf("failed to parse sunset time: %v", err)
	}

	localSunset := sunsetTime.Local().Format("15:04")
	return localSunset, nil
}

func GetSunrise(latitude, longitude float64) (string, error) {
	url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%f&lng=%f&formatted=0", latitude, longitude)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var data SunriseResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	if data.Status != "OK" {
		return "", fmt.Errorf("API response status: %s", data.Status)
	}

	sunriseTime, err := time.Parse(time.RFC3339, data.Results.Sunrise)
	if err != nil {
		return "", fmt.Errorf("failed to parse sunset time: %v", err)
	}

	localSunset := sunriseTime.Local().Format("15:04")
	return localSunset, nil
}
