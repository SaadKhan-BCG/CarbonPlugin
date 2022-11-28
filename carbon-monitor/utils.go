package main

import (
	"fmt"
	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/carbon_emissions"
	"github.com/SaadKhan-BCG/CarbonPlugin/carbon-monitor/container_stats"
	"github.com/docker/docker/client"
	"github.com/gosuri/uilive"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Outputs live updating console log of Carbon Data as stored in "totalCarbon"
// Note, outputs should be a list of keys in totalCarbon you wish to log
// Iterate over outputs instead of totalCarbon keys as map key ordering is not consistent so the output is harder to read
func outputLiveConsumption(writer *uilive.Writer, totalCarbon map[string]float64, outputs *[]string, name string) {
	for _, output := range *outputs {
		_, _ = fmt.Fprintf(writer, "%s: %s Carbon Consumed: %f\n", name, output, totalCarbon[output])
	}
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
		logrus.Fatal("Failed to Initialise Docker Client", err)
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

		container_stats.GetDockerStats(cli, containerPower)
		carbon_emissions.RefreshCarbonCache()

		logrus.Debug(containerPower)
		for container := range containerPower {
			logrus.Debug(containerPower[container])
			var wg sync.WaitGroup
			wg.Add(wgCount)
			for _, item := range *iterable {
				go computeFn(containerCarbon, container, containerPower[container], item, &wg)
			}
			wg.Wait()
		}
		logrus.Debug(containerCarbon)

		curTime = time.Now()
		diff = curTime.Sub(prevTime).Seconds()
		prevTime = time.Now()
		logrus.Debug(fmt.Sprintf("Time taken for iteration: %f Seconds", diff))
		logrus.Info(fmt.Sprintf("Total Time Spend: %f Seconds", curTime.Sub(startTime).Seconds()))

		for _, item := range *iterable {
			for container := range containerPower {
				totalCarbon[item] += containerCarbon[ContainerRegion{container, item}] * diff
			}
		}

		// Note, Live logging requires the loglevel be set to error as info logging gets in the way
		outputLiveConsumption(writer, totalCarbon, iterable, iterableName)
		logrus.Info(totalCarbon)

		// Empty the maps at the end of every iteration to prevent old reports staying
		containerPower = make(map[string]float64)
		containerCarbon = make(map[ContainerRegion]float64)
	}

	writer.Stop()
}
