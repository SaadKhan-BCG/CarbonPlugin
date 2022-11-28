package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	carbon_emissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	container_stats "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/container_stats"
	error_handler "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"

	"github.com/docker/docker/client"
	"github.com/gosuri/uilive"
	log "github.com/sirupsen/logrus"
)

// TODO put this somewhere better, ideally make it configurable via cli ie allow a --all option/ commar separated list of ones you care about
var regions = []string{
	"australiacentral",
	"australiacentral2",
	"australiaeast",
	"australiasoutheast",
	"canadacentral",
	"canadaeast",
	"centralus",
	"centraluseuap",
	"eastus",
	"eastus2",
	"eastus2euap",
	"northcentralus",
	"northeurope",
	"southcentralus",
	"uksouth",
	"ukwest",
	"westcentralus",
	"westus",
	"westus2",
	"westus3",
}

// TODO read from env var?
var delay = time.Second * 0

var mutex = &sync.Mutex{}

func ComputeCarbonConsumption(containerCarbon map[ContainerRegion]float64, container string, power float64, location string, wg *sync.WaitGroup) {
	defer wg.Done()
	carbon, err := carbon_emissions.GetCurrentCarbonEmissions(location)
	if err != nil {
		error_handler.StdErrorHandler(fmt.Sprintf("Failed to compute carbon for Container: %s Region: %s as could not get Carbon data", container, location), err)
	} else {
		carbonConsumed := power * carbon * 10 / 216 // Carbon is in gCo2/H converting here to mgCo2/S
		log.Debug(fmt.Sprintf("Location: %s Rating: %f Power: %f", location, carbon, power))
		mutex.Lock() // Map write operations are not thread safe and this function is called in parallel
		containerCarbon[ContainerRegion{container, location}] = carbonConsumed
		mutex.Unlock()
	}
}

var regionCount = len(regions)

type ContainerRegion struct {
	container string
	region    string
}

func fetch() {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal("Failed to Initialise Docker Client", err)
	}

	// TODO Improve this, storing far too many maps can be refactored
	containerPower := make(map[string]float64)
	containerCarbon := make(map[ContainerRegion]float64)
	totalCarbon := make(map[string]float64)
	startTime := time.Now()
	curTime := time.Now()
	prevTime := time.Now()
	diff := 0.0

	writer := uilive.New()
	writer.Start()

	for {
		time.Sleep(delay)

		container_stats.GetDockerStats(cli, containerPower)
		carbon_emissions.RefreshCarbonCache()

		log.Debug(containerPower)
		for container := range containerPower {
			log.Debug(containerPower[container])
			var wg sync.WaitGroup
			wg.Add(regionCount)
			for _, region := range regions {
				go ComputeCarbonConsumption(containerCarbon, container, containerPower[container], region, &wg)
			}
			wg.Wait()
		}
		log.Debug(containerCarbon)

		curTime = time.Now()
		diff = curTime.Sub(prevTime).Seconds()
		prevTime = time.Now()
		log.Debug(fmt.Sprintf("Time taken for iteration: %f Seconds", diff))
		log.Info(fmt.Sprintf("Total Time Spend: %f Seconds", curTime.Sub(startTime).Seconds()))
		for _, region := range regions {
			for container := range containerPower {
				totalCarbon[region] += containerCarbon[ContainerRegion{container, region}] * diff
			}
		}

		// Note, Live logging requires the loglevel be set to error as info logging gets in the way
		outputLiveConsumption(writer, totalCarbon)
		log.Info(totalCarbon)

		// Empty the maps at the end of every iteration to prevent old reports staying
		containerPower = make(map[string]float64)
		containerCarbon = make(map[ContainerRegion]float64)
	}

	writer.Stop()
}

func outputLiveConsumption(writer *uilive.Writer, totalCarbon map[string]float64) {
	for _, region := range regions {
		_, _ = fmt.Fprintf(writer, "Region: %s Carbon Consume: %f\n", region, totalCarbon[region])
	}
}

func main() {
	log.SetLevel(log.ErrorLevel)

	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	carbon_emissions.LoadSettings()

	fetch()
}
