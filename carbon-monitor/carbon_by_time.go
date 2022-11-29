package carbon

import (
	"fmt"
	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

var location string

func SetLocation(l string) {
	location = l
}

func ComputeCarbonConsumptionByTime(containerCarbon map[ContainerRegion]float64, container string, power float64, hour string, wg *sync.WaitGroup) {
	defer wg.Done()
	h, err := strconv.ParseInt(hour, 10, 64)
	if err != nil {
		error_handler.StdErrorHandler(fmt.Sprintf("Failed as could not parse hour value %s as int", hour), err)
		return
	}
	startTime := time.Now().AddDate(0, 0, -1)
	startTime = startTime.Add((time.Duration(h) - time.Duration(startTime.Hour())) * time.Hour)

	carbon, err := carbonemissions.GetCarbonEmissionsByTime(location, startTime)
	if err != nil {
		error_handler.StdErrorHandler(fmt.Sprintf("Failed to compute carbon for Container: %s Region: %s Hour: H%s as could not get Carbon data", container, location, hour), err)
	} else {
		carbonConsumed := power * carbon * 10 / 216 // Carbon is in gCo2/H converting here to mgCo2/S
		log.Debug(fmt.Sprintf("Location: %s Rating: %f Power: %f", location, carbon, power))
		mutex.Lock() // Map write operations are not thread safe and this function is called in parallel
		containerCarbon[ContainerRegion{container, hour}] = carbonConsumed
		mutex.Unlock()
	}
}
