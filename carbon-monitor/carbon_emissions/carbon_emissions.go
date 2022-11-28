package carbon_emissions

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"carbon-monitor/error_handler"
)

var baseUrl string

var carbonRegionCache map[string]float64
var carbonRegionTimeCache map[string]float64 // TODO Start using time cache too

var mutex = &sync.Mutex{}

type carbonAwareResponse struct {
	Rating   float64     `json:"rating"`
	Location interface{} `json:"location"`
	Time     interface{} `json:"time"`
	Duration interface{} `json:"duration"`
}

func getCarbonEmissions(location string, prevTime string, toTime string) (float64, error) {
	url := fmt.Sprintf("%s/emissions/bylocation?location=%s&time=%s&toTime=%s", baseUrl, location, prevTime, toTime)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data []carbonAwareResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err.Error())
	}
	return data[0].Rating, nil
}

func formatTimeAsString(time time.Time) string {
	return time.Format("2006-01-02T15:04")
}

func getCarbonEmissionsByTime(location string, utcTime time.Time) (float64, error) {
	toTime := formatTimeAsString(utcTime)
	prevTime := formatTimeAsString(utcTime.Add(-time.Minute))

	rating := carbonRegionCache[location]
	if rating > 0 {
		return rating, nil
	} else {
		rating, _ = getCarbonEmissions(location, prevTime, toTime)
		mutex.Lock()
		carbonRegionCache[location] = rating
		mutex.Unlock()
	}
	return rating, nil
}

func GetCurrentCarbonEmissions(location string) (float64, error) {
	rating, err := getCarbonEmissionsByTime(location, time.Now())
	if err != nil {
		error_handler.StdErrorHandler(fmt.Sprintf("Failure fetching emission data for Region: %s", location), err)
		return 0, err
	} else {
		log.Debug(fmt.Sprintf("Location: %s Rating: %f", location, rating))
		return rating, nil
	}
}

func RefreshCarbonCache() {
	carbonRegionCache = make(map[string]float64)
	carbonRegionTimeCache = make(map[string]float64)
}

func LoadSettings() {
	RefreshCarbonCache()

	url := os.Getenv("CARBON_SDK_URL")
	host := os.Getenv("CARBON_SDK_HOST")
	port := os.Getenv("CARBON_SDK_PORT")
	if url != "" {
		baseUrl = url
	} else {
		if host == "" || port == "" {
			log.Fatal("Error loading env variables, please set either CARBON_SDK_URL or CARBON_SDK_HOST and CARBON_SDK_PORT")
		}
		baseUrl = fmt.Sprintf("http://%s:%s", host, port)
	}
}
