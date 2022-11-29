package carbon

import (
	"fmt"
	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
	log "github.com/sirupsen/logrus"
	"sync"
)

func ComputeCurrentCarbonConsumption(containerCarbon map[ContainerRegion]float64, container string, power float64, location string, wg *sync.WaitGroup) {
	defer wg.Done()
	carbon, err := carbonemissions.GetCurrentCarbonEmissions(location)
	if err != nil {
		error_handler.StdErrorHandler(fmt.Sprintf("Failed fetching emissions data for Container: %s Region: %s ", container, location), err)
	} else {
		carbonConsumed := power * carbon * 10 / 216 // Carbon is in gCo2/H converting here to mgCo2/S
		log.Debug(fmt.Sprintf("Location: %s Rating: %f Power: %f", location, carbon, power))
		mutex.Lock() // Map write operations are not thread safe and this function is called in parallel
		containerCarbon[ContainerRegion{container, location}] = carbonConsumed
		mutex.Unlock()
	}
}
