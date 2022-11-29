package carbon

import (
	"fmt"
	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
	"sync"
)

func ComputeCurrentCarbonConsumption(containerCarbon map[ContainerRegion]float64, container string, power float64, location string, wg *sync.WaitGroup) {
	defer wg.Done()
	carbon, err := carbonemissions.GetCurrentCarbonEmissions(location)
	if err != nil {
		error_handler.StdErrorHandler(fmt.Sprintf("Failed fetching emissions data for Container: %s Region: %s ", container, location), err)
	} else {
		computeAndUpdateCarbonConsumption(containerCarbon, container, power, location, carbon)
	}
}
