package main

import (
	"carbon-monitor/carbon_emissions"
	"carbon-monitor/container_stats"
	"fmt"
	"os"
	"time"
)

func main() {
	os.Setenv("CARBON_SDK_URL", "https://carbon-aware-api.azurewebsites.net")
	carbon_emissions.LoadSettings()

	carbon_emissions.GetCarbonEmissionsByTime("uksouth", time.Now())
	fmt.Println(carbon_emissions.CarbonRegionCache["foo"])
	carbon := carbon_emissions.GetCarbonTest()
	fmt.Println(carbon_emissions.CarbonRegionCache["foo"])
	fmt.Println(carbon)

	for {
		stats := container_stats.GetDockerStats()
		fmt.Println(stats)
	}
}
