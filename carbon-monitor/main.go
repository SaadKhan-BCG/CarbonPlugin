package main

import (
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

func ComputeCarbonConsumption(containerCarbon map[ContainerRegion]float64, container string, power float64, location string, wg *sync.WaitGroup) {
	defer wg.Done()
	carbon, _ := carbon_emissions.GetCurrentCarbonEmissions(location)
	containerCarbon[ContainerRegion{container, location}] = power * carbon
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

	for {
		log.Info(container_stats.GetDockerStats(cli, containerPower))
		for container := range containerPower {
			log.Debug(containerPower[container])
			var wg sync.WaitGroup
			wg.Add(regionCount)
			for _, region := range regions {
				go ComputeCarbonConsumption(containerCarbon, container, containerPower[container], region, &wg)
			}
			wg.Wait()
		}
		log.Info(containerCarbon)
	}
}

func main() {
	log.SetLevel(log.InfoLevel)

	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	carbon_emissions.LoadSettings()

	carbon_emissions.GetCarbonEmissionsByTime("uksouth", time.Now())

	fetch()
}
