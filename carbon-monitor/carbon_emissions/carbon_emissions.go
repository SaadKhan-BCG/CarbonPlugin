package carbon_emissions

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"time"
)

var BaseUrl string

// TODO Use this cache to reduce external queries
var CarbonRegionCache map[string]float64
var CarbonRegionTimeCache map[string]float64

type CarbonAwareResponse struct {
	Rating   float64     `json:"rating"`
	Location interface{} `json:"location"`
	Time     interface{} `json:"time"`
	Duration interface{} `json:"duration"`
}

func getCarbonEmissions(location string, prevTime string, toTime string) (float64, error) {
	url := fmt.Sprintf("%s/emissions/bylocation?location=%s&time=%s&toTime=%s", BaseUrl, location, prevTime, toTime)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data []CarbonAwareResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Results: %v\n", data)
	return data[0].Rating, nil
}

func formatTimeAsString(time time.Time) string {
	return time.Format("2006-01-02T15:04")
}

func GetCarbonEmissionsByTime(location string, utcTime time.Time) (float64, error) {
	toTime := formatTimeAsString(utcTime)
	prevTime := formatTimeAsString(utcTime.Add(-time.Minute))

	rating, _ := getCarbonEmissions(location, prevTime, toTime)
	return rating, nil
}

func GetCurrentCarbonEmissions(location string) (float64, error) {
	rating, _ := GetCarbonEmissionsByTime(location, time.Now())
	return rating, nil
}

func LoadSettings() {
	CarbonRegionCache = make(map[string]float64)
	CarbonRegionTimeCache = make(map[string]float64)

	url := os.Getenv("CARBON_SDK_URL")
	host := os.Getenv("CARBON_SDK_HOST")
	port := os.Getenv("CARBON_SDK_PORT")
	if url != "" {
		BaseUrl = url
	} else {
		if host == "" || port == "" {
			log.Fatal("Error loading env variables, please set either CARBON_SDK_URL or CARBON_SDK_HOST and CARBON_SDK_PORT")
		}
		BaseUrl = fmt.Sprintf("http://%s:%s", host, port)
	}
}
