package carbon

import (
	"fmt"
	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	errorhandler "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
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
		errorhandler.StdErrorHandler(fmt.Sprintf("Failed as could not parse hour value %s as int", hour), err)
		return
	}
	startTime := time.Now().AddDate(0, 0, -1) // Take the day before as the reference since today's values are not yet fully available
	startTime = startTime.Add((time.Duration(h) - time.Duration(startTime.Hour())) * time.Hour)

	carbon, err := carbonemissions.GetCarbonEmissionsByTime(location, startTime)
	if err != nil {
		errorhandler.StdErrorHandler(fmt.Sprintf("Failed fetching emissions data for Container: %s Region: %s Hour: H%s ", container, location, hour), err)
	} else {
		computeAndUpdateCarbonConsumption(containerCarbon, container, power, hour, carbon)
	}
}
