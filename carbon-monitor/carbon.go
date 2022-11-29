package carbon

import (
	"os"
	"sync"
	"time"

	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
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

func init() {
	log.SetLevel(log.ErrorLevel)

	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	carbonemissions.LoadSettings()
}

//func main() {
//	//defaultRegions = []string{"ukwest", "uksouth", "australiacentral"}
//	//RegionMode(&defaultRegions)
//	//TimeMode("uksouth")
//	//GraphMode("uksouth")
//}
