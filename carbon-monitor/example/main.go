package main

import (
	carbon "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor"
	"os"
)

func main() {
	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	//carbon.RegionMode(&[]string{"uksouth", "foo"})
	//carbon.RegionMode(&[]string{"westus"})
	carbon.GraphMode("westus")
	//carbon.TimeMode("westus")
	//regions := strings.Join(carbon.ListValidRegions(), "\n")
	//log.SetLevel(log.InfoLevel)
	//log.Println("Available regions for measuring carbon consumption:")
	//log.Println(regions)
}
