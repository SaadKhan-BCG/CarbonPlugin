package carbon

import (
	"fmt"
	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

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

var mutex = &sync.Mutex{}

var wgCount int

type ContainerRegion struct {
	container string
	item      string // Some second dimension to group data on, ie region, time
}

func RegionMode(regions *[]string) {
	if len(*regions) < 1 {
		regions = &defaultRegions
	}
	OutputTotalCarbon("Region", regions, ComputeCurrentCarbonConsumption)
}

func TimeMode(region string) {
	SetLocation(region)
	OutputTotalCarbon("Hour", &timeZones, ComputeCarbonConsumptionByTime)
}

func GraphMode(region string) {
	asciPlot(region)
}

func ListValidRegions() []string {
	return defaultRegions
}

func LoadEnvVars() {
	// TODO think about a better way to avoid this. This line makes it impossible to configure nicely
	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	carbonUrl := os.Getenv("CARBON_SDK_URL")

	if len(carbonUrl) < 1 {
		host := GetOrElsEnvVars("CARBON_SDK_HOST", "localhost")
		port := GetOrElsEnvVars("CARBON_SDK_PORT", "8080")
		log.Info(fmt.Sprintf("CarbonAwareSDK: CARBON_SDK_URL not found defaulting to http://%s:%s", host, port))
	} else {
		log.Info(fmt.Sprintf("Using Carbon Aware SDK at %s", carbonUrl))
	}
}

func init() {
	log.SetLevel(log.ErrorLevel)
	LoadEnvVars()
	carbonemissions.LoadSettings()
}
