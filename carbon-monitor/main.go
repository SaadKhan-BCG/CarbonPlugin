package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"carbon-monitor/carbon_emissions"
	"carbon-monitor/container_stats"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

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
var mutex = &sync.Mutex{}

func ComputeCarbonConsumption(containerCarbon map[ContainerRegion]float64, container string, power float64, location string, wg *sync.WaitGroup) {
	defer wg.Done()
	carbon, _ := carbon_emissions.GetCurrentCarbonEmissions(location)
	carbonConsumed := power * carbon * 10 / 36 // Carbon is in gCo2/H converting here to mgCo2/S
	mutex.Lock()
	containerCarbon[ContainerRegion{container, location}] = carbonConsumed
	mutex.Unlock()
}

var regionCount = len(regions)

type ContainerRegion struct {
	container string
	region    string
}

func fetch() {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	containerPower := make(map[string]float64)
	containerCarbon := make(map[ContainerRegion]float64)
	totalCarbon := make(map[string]float64)
	curTime := time.Now()
	prevTime := time.Now()
	diff := 0.0

	for {
		container_stats.GetDockerStats(cli, containerPower)
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
		for _, region := range regions {
			for container := range containerPower {
				totalCarbon[region] += containerCarbon[ContainerRegion{container, region}] * diff
			}
		}

		log.Info(totalCarbon)

		// Empty the maps at the end of every iteration to prevent old reports staying
		containerPower = make(map[string]float64)
		containerCarbon = make(map[ContainerRegion]float64)
	}
}

func main() {
	log.SetLevel(log.InfoLevel)

	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	carbon_emissions.LoadSettings()

	carbon_emissions.GetCarbonEmissionsByTime("uksouth", time.Now())

	fetch()
}
