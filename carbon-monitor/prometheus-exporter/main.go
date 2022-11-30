package main

import (
	"flag"
	"fmt"
	carbon "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor"
	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	containerstats "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/container_stats"
	errorhandler "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//var Regions = carbon.ListValidRegions()

var Regions = []string{"westus"} // Temp while only using free account
var RegionLen = len(Regions)

func recordMetrics() {
	go func() {
		cli, err := client.NewEnvClient()
		if err != nil {
			log.Fatal("Failed to Initialise Docker Client", err)
		}
		for {
			containerPower := make(map[string]float64)
			containerstats.GetDockerStats(cli, containerPower)
			for container := range containerPower {
				updateContainerMetrics(container, containerPower[container])
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func updateContainerMetrics(containerName string, power float64) {
	powerUsage.WithLabelValues(containerName).Set(power)
	log.Info(fmt.Sprintf("Power for Container: %s Watts: %f", containerName, power))
	var wg sync.WaitGroup
	wg.Add(RegionLen)
	for _, region := range Regions {
		go updateContainerRegionMetrics(containerName, power, region, &wg)
	}
	wg.Wait()

	//var wg2 sync.WaitGroup
	//wg2.Add(TimeRegionsLen)
	for _, region := range TimeRegions {
		for i := 1; i < 6; i++ {
			updateContainerTimeMetrics(containerName, power, region, i*4)
		}
	}
	//wg2.Wait()
}

func updateContainerTimeMetrics(containerName string, power float64, region string, hour int) {
	//defer wg.Done()
	startTime := time.Now().AddDate(0, 0, -1)
	startTime = startTime.Add((time.Duration(hour) - time.Duration(startTime.Hour())) * time.Hour)

	carbonRate, err := carbonemissions.GetCarbonEmissionsByTime(region, startTime)
	if err != nil {
		errorhandler.StdErrorHandler(fmt.Sprintf("Failed fetching emissions data for Container: %s Region: %s ", containerName, region), err)
	} else {
		carbonConsumptionTime.WithLabelValues(containerName, region, fmt.Sprintf("%d", hour)).Set(power * carbonRate)
	}
}

func updateContainerRegionMetrics(containerName string, power float64, region string, wg *sync.WaitGroup) {
	defer wg.Done()
	carbonRate, err := carbonemissions.GetCurrentCarbonEmissions(region)
	if err != nil {
		errorhandler.StdErrorHandler(fmt.Sprintf("Failed fetching emissions data for Container: %s Region: %s ", containerName, region), err)
	} else {
		carbonConsumption.WithLabelValues(containerName, region).Set(power * carbonRate)
	}
}

var powerUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "power_usage",
	Help: "TODO",
}, []string{"container_name"})

var carbonConsumption = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "carbon_consumption",
	Help: "TODO",
}, []string{"container_name", "region"})

var carbonConsumptionTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "carbon_consumption_time",
	Help: "TODO",
}, []string{"container_name", "region", "time"})

func registerMetrics() {
	prometheus.MustRegister(powerUsage)
	prometheus.MustRegister(carbonConsumption)
	prometheus.MustRegister(carbonConsumptionTime)
}

var timeRegionStr string
var TimeRegions []string
var TimeRegionsLen int

func readFlags() {
	flag.StringVar(&timeRegionStr, "timeRegions", "", "Regions to collect and export time data on")
	flag.Parse()

	if len(timeRegionStr) == 0 {
		timeRegionStr = os.Getenv("TIME_REGIONS") // If not set via cli check env var
		if len(timeRegionStr) > 0 {
			TimeRegions = strings.Split(timeRegionStr, ",")
		} else {
			TimeRegions = []string{} // Parse no values as empty list to prevent querying for region ""
		}

	} else {
		TimeRegions = strings.Split(timeRegionStr, ",")
	}
	TimeRegionsLen = len(TimeRegions) * 6 // 6 time points per region ie one every 4 hours
}

func startPrometheusMetrics() {
	registerMetrics()
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func main() {
	log.SetLevel(log.InfoLevel)

	carbon.LoadEnvVars()
	carbonemissions.LoadSettings()

	readFlags()

	startPrometheusMetrics()
}
