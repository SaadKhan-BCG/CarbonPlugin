package main

import (
	"fmt"
	carbon "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor"
	carbonemissions "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	containerstats "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/container_stats"
	errorhandler "github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/error_handler"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Regions = carbon.ListValidRegions()
var RegionLen = len(Regions)

func recordMetrics() {
	go func() {
		cli, err := client.NewEnvClient()
		if err != nil {
			log.Fatal("Failed to Initialise Docker Client", err)
		}
		containerPower := make(map[string]float64)
		for {
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

func main() {
	log.SetLevel(log.InfoLevel)

	registerMetrics()
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
