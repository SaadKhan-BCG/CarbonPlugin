package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/gosuri/uilive"
	"os"
	"sync"
	"time"

	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	containerstats "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/container_stats"
	log "github.com/sirupsen/logrus"
)

// TODO put this somewhere better, ideally make it configurable via cli ie allow a --all option/ commar separated list of ones you care about
var defaultRegions = []string{
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

var timeZones = []string{"0", "4", "8", "12", "16", "20"}

// TODO read from env var?
var delay = time.Second * 0

var mutex = &sync.Mutex{}

var wgCount int

type ContainerRegion struct {
	container string
	item      string // Some second dimension to group data on, ie region, time
}

func main() {
	//defaultRegions = []string{"ukwest", "uksouth", "australiacentral"}
	RegionMode(&defaultRegions)
	//TimeMode("uksouth")
}

func RegionMode(regions *[]string) {
	log.SetLevel(log.ErrorLevel)

	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	carbonemissions.LoadSettings()

	OutputTotalCarbon("Region", regions, ComputeCurrentCarbonConsumption)
}

func TimeMode(region string) {
	log.SetLevel(log.ErrorLevel)

	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	carbonemissions.LoadSettings()

	SetLocation(region)
	OutputTotalCarbon("Hour", &timeZones, ComputeCarbonConsumptionByTime)
}

/*
		OutputTotalCarbon Generic function to Output Total Carbon split by some iterable, ie region or time ranges
		@param: iterableName Name of the variable you are iterating over, only used in logging
		@param: Iterable The variable you iterate over e.g. regions, time zones. Note we will refer to an instance in this as "item"
		@param computeFn: Function which takes as input containerCarbon -> map(containerName, item) -> carbon consumed, containerName, power (consumed by container), item.
	                      Should update containerCarbon with the correct carbon value for containerName, item tuple
*/
func OutputTotalCarbon(iterableName string, iterable *[]string, computeFn func(map[ContainerRegion]float64, string, float64, string, *sync.WaitGroup)) {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal("Failed to Initialise Docker Client", err)
	}

	wgCount = len(*iterable)

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

		containerstats.GetDockerStats(cli, containerPower)
		carbonemissions.RefreshCarbonCache()

		log.Debug(containerPower)
		for container := range containerPower {
			log.Debug(containerPower[container])
			var wg sync.WaitGroup
			wg.Add(wgCount)
			for _, item := range *iterable {
				go computeFn(containerCarbon, container, containerPower[container], item, &wg)
			}
			wg.Wait()
		}
		log.Debug(containerCarbon)

		curTime = time.Now()
		diff = curTime.Sub(prevTime).Seconds()
		prevTime = time.Now()
		log.Debug(fmt.Sprintf("Time taken for iteration: %f Seconds", diff))
		log.Info(fmt.Sprintf("Total Time Spend: %f Seconds", curTime.Sub(startTime).Seconds()))
		for _, item := range *iterable {
			for container := range containerPower {
				totalCarbon[item] += containerCarbon[ContainerRegion{container, item}] * diff
			}
		}

		// Note, Live logging requires the loglevel be set to error as info logging gets in the way
		outputLiveConsumption(writer, totalCarbon, iterable, iterableName)
		log.Info(totalCarbon)

		// Empty the maps at the end of every iteration to prevent old reports staying
		containerPower = make(map[string]float64)
		containerCarbon = make(map[ContainerRegion]float64)
	}

	writer.Stop()
}

// Outputs live updating console log of Carbon Data as stored in "totalCarbon"
// Note, outputs should be a list of keys in totalCarbon you wish to log
// Iterate over outputs instead of totalCarbon keys as map key ordering is not consistent so the output is harder to read
func outputLiveConsumption(writer *uilive.Writer, totalCarbon map[string]float64, outputs *[]string, name string) {
	for _, output := range *outputs {
		_, _ = fmt.Fprintf(writer, "%s: %s Carbon Consumed: %f\n", name, output, totalCarbon[output])
	}
}
