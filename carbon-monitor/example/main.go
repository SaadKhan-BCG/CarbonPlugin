package main

import (
	carbon "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor"
)

func main() {
	//carbon.RegionMode(&[]string{"uksouth", "foo"})
	//carbon.RegionMode(&[]string{"westus"})
	carbon.GraphMode("westus")
	//carbon.TimeMode("westus")
	//regions := strings.Join(carbon.ListValidRegions(), "\n")
	//log.SetLevel(log.InfoLevel)
	//log.Println("Available regions for measuring carbon consumption:")
	//log.Println(regions)
}
